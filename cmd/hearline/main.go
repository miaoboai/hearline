// Command hearline is the v0.1 desktop loop: pick audio, run SenseVoice-Small
// Q8 on CPU, and show subtitles. Without a display, `demo` and `transcribe`
// decode a WAV and write the lines.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/asr/sensevoice"
	"github.com/miaoboai/hearline/internal/audio"
	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/engine"
	"github.com/miaoboai/hearline/internal/llm"
	"github.com/miaoboai/hearline/internal/pipeline"
	"github.com/miaoboai/hearline/internal/ui"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		os.Exit(runGUI(nil))
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var code int
	switch cmd {
	case "gui":
		code = runGUI(args)
	case "demo":
		code = runDemo(args)
	case "transcribe":
		code = runTranscribe(args)
	case "download-model":
		code = runDownload(args)
	case "sources":
		code = runSources()
	case "config":
		code = runConfig(args)
	case "version", "-version", "--version":
		fmt.Println("hearline", version)
		return
	case "help", "-h", "--help":
		usage()
		return
	default:
		if strings.HasPrefix(cmd, "-") {
			code = runGUI(os.Args[1:])
			break
		}
		usage()
		code = 2
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Hearline %s — 本机听写字幕（SenseVoice-Small Q8，CPU）

  hearline                       打开字幕窗口
  hearline download-model        下载 SenseVoice-Small int8（约 229MB）
  hearline demo                  用官方 zh.wav / en.wav 跑通识别
  hearline transcribe --wav FILE 识别一个 WAV，把字幕打到标准输出
  hearline sources               列出 Pulse/PipeWire 音源
  hearline config init           写出 ~/.config/hearline/config.json（0600）
  hearline version

无显示器时用 demo 或 transcribe。LLM 只出现在配置里，字幕环不调用它。
`, version)
}

func runGUI(args []string) int {
	fs := flag.NewFlagSet("gui", flag.ExitOnError)
	model := fs.String("model-dir", "", "SenseVoice 模型目录")
	engineName := fs.String("engine", "", "sensevoice、mock 或 qwen3")
	threads := fs.Int("threads", 0, "CPU 线程，0 为自动，最大 4")
	_ = fs.Parse(args)
	cfg, err := load(*model, *engineName, *threads)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	err = ui.Run(ui.Options{
		LoadRecognizer: func() (asr.Recognizer, error) { return engine.New(cfg.ASR) },
		ListSources:    audio.List,
		Capture:        audio.Capture,
		LLMStatus:      llm.Describe(cfg.LLM),
		Chunk:          chunkOpts(cfg, true),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runDemo(args []string) int {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)
	model := fs.String("model-dir", "", "模型目录；缺省则下载到用户缓存")
	out := fs.String("out", "", "把字幕行写入这个文件")
	shot := fs.String("screenshot", "", "把字幕界面画成 PNG")
	threads := fs.Int("threads", 0, "CPU 线程，0 为自动，最大 4")
	_ = fs.Parse(args)

	dir := *model
	if _, err := sensevoice.Resolve(dir); err != nil {
		dest := dir
		if dest == "" {
			cache, cerr := os.UserCacheDir()
			if cerr != nil {
				fmt.Fprintln(os.Stderr, cerr)
				return 1
			}
			dest = cache + "/hearline/sensevoice-small-q8"
		}
		fmt.Fprintf(os.Stderr, "模型不在本地，开始下载到 %s\n", dest)
		if err := sensevoice.Download(context.Background(), dest); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		dir = dest
	}
	paths, err := sensevoice.Resolve(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg, err := load(paths.Dir, "sensevoice", *threads)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Fprintf(os.Stderr, "loading recognizer %s\n", cfg.ASR.Engine)
	rec, err := engine.New(cfg.ASR)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer rec.Close()
	var all []pipeline.Line
	var texts []string
	for _, name := range []string{"zh.wav", "en.wav"} {
		wav := sensevoice.TestWav(paths.Dir, name)
		fmt.Fprintf(os.Stderr, "transcribe %s\n", wav)
		lines, text, err := decodeFile(rec, wav, chunkOpts(cfg, false))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("# %s\n", name)
		for _, line := range lines {
			if line.Partial {
				continue
			}
			fmt.Println(line.String())
		}
		all = append(all, lines...)
		texts = append(texts, text)
		fmt.Println()
	}
	if *out != "" {
		var b strings.Builder
		for _, line := range finals(all) {
			fmt.Fprintln(&b, line.String())
		}
		if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", *out)
	}
	if *shot != "" {
		m := ui.Snapshot(finals(all), "官方测试 WAV（zh + en）", "SenseVoice-Small Q8 · CPU · "+fmt.Sprint(cfg.ASR.Threads())+" threads", llm.Describe(cfg.LLM))
		if err := ui.WritePNG(m, *shot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", *shot)
	}
	zh, en := "", ""
	if len(texts) == 2 {
		zh, en = texts[0], strings.ToLower(texts[1])
	}
	if !strings.Contains(zh, "点") || !strings.Contains(zh, "下午") {
		fmt.Fprintf(os.Stderr, "zh transcript did not match SenseVoice fixture: %q\n", zh)
		return 1
	}
	if !strings.Contains(en, "boy") && !strings.Contains(en, "tribal") && !strings.Contains(en, "chieftain") {
		fmt.Fprintf(os.Stderr, "en transcript did not match SenseVoice fixture: %q\n", en)
		return 1
	}
	fmt.Fprintln(os.Stderr, "demo ok")
	return 0
}

func runTranscribe(args []string) int {
	fs := flag.NewFlagSet("transcribe", flag.ExitOnError)
	wav := fs.String("wav", "", "WAV 路径")
	model := fs.String("model-dir", "", "模型目录")
	out := fs.String("out", "", "字幕输出文件")
	shot := fs.String("screenshot", "", "字幕界面 PNG")
	engineName := fs.String("engine", "", "sensevoice 或 mock")
	threads := fs.Int("threads", 0, "CPU 线程，0 为自动，最大 4")
	maxChunk := fs.Float64("max-chunk", 8, "单次解码最长秒数")
	_ = fs.Parse(args)
	if *wav == "" {
		fmt.Fprintln(os.Stderr, "transcribe 需要 --wav")
		return 2
	}
	cfg, err := load(*model, *engineName, *threads)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *maxChunk > 0 {
		cfg.ASR.ChunkSec = *maxChunk
	}
	lines, text, err := transcribeFile(cfg, *wav)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	for _, line := range lines {
		if !line.Partial {
			fmt.Println(line.String())
		}
	}
	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "no subtitle text")
		return 1
	}
	if *out != "" {
		if err := os.WriteFile(*out, []byte(text+"\n"), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if *shot != "" {
		m := ui.Snapshot(finals(lines), *wav, cfg.ASR.Engine+" · CPU", llm.Describe(cfg.LLM))
		if err := ui.WritePNG(m, *shot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

func transcribeFile(cfg config.Config, wav string) ([]pipeline.Line, string, error) {
	fmt.Fprintf(os.Stderr, "loading recognizer %s\n", cfg.ASR.Engine)
	rec, err := engine.New(cfg.ASR)
	if err != nil {
		return nil, "", err
	}
	defer rec.Close()
	return decodeFile(rec, wav, chunkOpts(cfg, false))
}

func decodeFile(rec asr.Recognizer, wav string, opt pipeline.Options) ([]pipeline.Line, string, error) {
	samples, err := audio.ReadMono16k(wav)
	if err != nil {
		return nil, "", err
	}
	fmt.Fprintf(os.Stderr, "%s: %d samples, %.2fs\n", wav, len(samples), float64(len(samples))/audio.SampleRate)
	start := time.Now()
	lines, err := pipeline.TranscribePCM(context.Background(), rec, samples, audio.SampleRate, opt)
	if err != nil {
		return nil, "", err
	}
	elapsed := time.Since(start)
	audioSec := float64(len(samples)) / audio.SampleRate
	rtf := 0.0
	if audioSec > 0 {
		rtf = elapsed.Seconds() / audioSec
	}
	fmt.Fprintf(os.Stderr, "decoded in %s, audio %.2fs, RTF %.3f, lines %d\n", elapsed.Round(time.Millisecond), audioSec, rtf, len(finals(lines)))
	var b strings.Builder
	for _, line := range finals(lines) {
		fmt.Fprintf(&b, "%s\n", line.Text)
	}
	return lines, strings.TrimSpace(b.String()), nil
}

func finals(lines []pipeline.Line) []pipeline.Line {
	out := make([]pipeline.Line, 0, len(lines))
	for _, line := range lines {
		if !line.Partial && line.Text != "" {
			out = append(out, line)
		}
	}
	return out
}

func runDownload(args []string) int {
	fs := flag.NewFlagSet("download-model", flag.ExitOnError)
	dir := fs.String("dir", "", "解压目录，默认用户缓存")
	_ = fs.Parse(args)
	if err := sensevoice.Download(context.Background(), *dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	paths, err := sensevoice.Resolve(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(paths.Dir)
	return 0
}

func runSources() int {
	list, err := audio.List()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if len(list) == 0 {
		fmt.Println("(no live sources; WAV files still work)")
		return 0
	}
	for _, s := range list {
		fmt.Printf("%s\t%s\t%s\n", s.Kind, s.ID, s.Name)
	}
	return 0
}

func runConfig(args []string) int {
	if len(args) == 0 || args[0] != "init" {
		fmt.Fprintln(os.Stderr, "usage: hearline config init")
		return 2
	}
	cfg := config.Default()
	path := config.DefaultPath()
	if err := cfg.Save(path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println(path)
	return 0
}

func load(modelDir, engineName string, threads int) (config.Config, error) {
	cfg, err := config.Load("")
	if err != nil {
		return config.Config{}, err
	}
	if modelDir != "" {
		cfg.ASR.ModelDir = modelDir
	}
	if engineName != "" {
		cfg.ASR.Engine = engineName
	}
	if threads > 0 {
		cfg.ASR.NumThreads = threads
	}
	if err := cfg.Validate(); err != nil {
		return config.Config{}, err
	}
	return cfg, nil
}

func chunkOpts(cfg config.Config, live bool) pipeline.Options {
	opt := pipeline.Options{MaxChunk: time.Duration(cfg.ASR.ChunkSec * float64(time.Second))}
	if live {
		opt.PartialEvery = 2 * time.Second
	}
	return opt
}
