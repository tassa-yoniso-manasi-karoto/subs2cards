package cmd

import (
	"os"
	"strings"
	"path"
	"path/filepath"
	"time"
	"runtime"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/k0kubun/pp"
	"github.com/gookit/color"
	
	"github.com/tassa-yoniso-manasi-karoto/langkit/pkg/media"
)


type Task struct {
	Log                  zerolog.Logger
	Meta                 Meta
	OriginalLang         string // FIXME what for?
	Langs                []string
	RefLangs             []Lang
	Targ                 Lang
	SeparationLib        string
	STT                  string
	TargetChan           int
	UseAudiotrack        int
	TimeoutSTT           int
	TimeoutSep           int
	Offset               time.Duration
	WantDubs             bool
	DubsOnly             bool
	IsCCorDubs           bool
	TargSubFile          string
	RefSubFile           string
	MediaSourceFile      string
	FieldSep             string // defaults to "\t"
	OutputFileExtension  string // defaults to ".tsv" for "\t" and ".csv", otherwise
}


type Meta struct {
	FFmpeg string
	MediaInfo MediaInfo
	Runtime string
}

func DefaultTask(cmd *cobra.Command) (*Task) {
	var tsk Task
	if tsk.FieldSep == "" {
		tsk.FieldSep = "\t"
	}

	if tsk.OutputFileExtension == "" {
		switch tsk.FieldSep {
		case "\t":
			tsk.OutputFileExtension = ".tsv"
		default:
			tsk.OutputFileExtension = ".csv"
		}
	}
	bin := "ffmpeg"
	if runtime.GOOS == "windows" {
		bin = "ffmpeg.exe"
	}
	ex, err := os.Executable()
	if err != nil {
		logger.Warn().Err(err).Msg("failed to access directory where langkit is"+
			"FFmpeg path must be specified manually")
	}
	local := path.Join(filepath.Dir(ex), "bin", bin)
	if _, err := os.Stat(local); err == nil {
		media.FFmpegPath = local
	} else {
		media.FFmpegPath = bin
	}
	if cmd.Flags().Changed("ffmpeg") {
		media.FFmpegPath, _ = cmd.Flags().GetString("ffmpeg")
	}
	tsk.Meta.FFmpeg, err = getFFmpegVersion(media.FFmpegPath)
	if err != nil {
		logger.Fatal().Err(err). Msg("failed to access FFmpeg binary")
	}
	tsk.Meta.Runtime = getRuntimeInfo()
	targetChan, _ := cmd.Flags().GetInt("chan")
	audiotrack, _ := cmd.Flags().GetInt("a")
	//CC, _         := cmd.Flags().GetBool("cc")
	tsk = Task{
		Log:                  logger,
		Langs:                langs,
		TargetChan:           targetChan,
		//IsCC:                 CC,
		UseAudiotrack:        audiotrack-1,
		FieldSep:             "\t",
		OutputFileExtension:  "tsv",
	}
	return &tsk
}


func (tsk *Task) routing() {
	mediafile := tsk.MediaSourceFile
	media, err := os.Stat(mediafile)
	if err != nil {
		logger.Fatal().Err(err).Msg("can't access passed media file/directory")
	}
	if !media.IsDir() {
		tsk.Execute()
	} else {
		err = filepath.Walk(mediafile, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				tsk.Log.Fatal().Err(err).Msg("error during recursive exploration of passed directory")
			}
			if info.IsDir() && strings.HasSuffix(info.Name(), ".media") {
				return filepath.SkipDir
			}
			filename := filepath.Base(path)
			if !strings.HasSuffix(path, ".mp4") && !strings.HasSuffix(filename, ".mkv")  {
				return nil
			}
			tsk.RefSubFile = ""
			tsk.TargSubFile = ""
			tsk.MediaSourceFile = path
			tsk.Execute() // TODO go tsk.Execute()?
			return nil
		})
	}
}

func getFFmpegVersion(FFmpegPath string) (string, error) {
	cmd := exec.Command(FFmpegPath, "-version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run ffmpeg: %v", err)
	}

	// Convert output to a string
	output := out.String()

	// Define a regular expression to extract the version and commit hash
	// Example output: "ffmpeg version 4.3.2 Copyright (c) 2000-2021 the FFmpeg developers"
	re := regexp.MustCompile(`ffmpeg version (\S+)`)
	match := re.FindStringSubmatch(output)

	if len(match) < 2 {
		return "", fmt.Errorf("failed to extract ffmpeg version from output")
	}

	// Return the version found in the output
	return match[1], nil
}


func getRuntimeInfo() string {
	var sb strings.Builder

	// Store Go version
	sb.WriteString(fmt.Sprintf("\nGo version: %s\n", runtime.Version()))

	// Store OS and architecture information
	sb.WriteString(fmt.Sprintf("OS: %s\nArchitecture: %s\n", runtime.GOOS, runtime.GOARCH))

	// Store number of CPUs
	sb.WriteString(fmt.Sprintf("Number of CPUs: %d\n", runtime.NumCPU()))

	// Store number of Goroutines
	sb.WriteString(fmt.Sprintf("Number of Goroutines: %d\n", runtime.NumGoroutine()))

	// Store memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	sb.WriteString(fmt.Sprintf("Memory Allocated: %d bytes\n", memStats.Alloc))
	sb.WriteString(fmt.Sprintf("Memory Total Allocated: %d bytes\n", memStats.TotalAlloc))
	sb.WriteString(fmt.Sprintf("Memory System: %d bytes\n", memStats.Sys))
	/*sb.WriteString(fmt.Sprintf("Garbage Collection Cycles: %d\n", memStats.NumGC))

	// Store time since program started
	startTime := time.Now()
	sb.WriteString(fmt.Sprintf("Current time: %s\n", startTime.Format(time.RFC1123)))

	// Store process information
	pid := os.Getpid()
	sb.WriteString(fmt.Sprintf("Process ID: %d\n", pid))
	sb.WriteString(fmt.Sprintf("Parent Process ID: %d\n", os.Getppid()))

	// Store host information
	hostname, err := os.Hostname()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error retrieving hostname: %s\n", err))
	} else {
		sb.WriteString(fmt.Sprintf("Hostname: %s\n", hostname))
	}

	// Store environment variables (can filter sensitive variables if necessary)
	envVars := os.Environ()
	sb.WriteString("Environment Variables:\n")
	for _, env := range envVars {
		sb.WriteString(fmt.Sprintf("%s\n", env))
	}*/
	return sb.String()
}





func placeholder2345634567() {
	color.Redln(" 𝒻*** 𝓎ℴ𝓊 𝒸ℴ𝓂𝓅𝒾𝓁ℯ𝓇")
	pp.Println("𝓯*** 𝔂𝓸𝓾 𝓬𝓸𝓶𝓹𝓲𝓵𝓮𝓻")
}
