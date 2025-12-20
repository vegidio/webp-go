package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"github.com/vegidio/webp-go"
)

func main() {
	var quality uint

	cmd := &cli.Command{
		Name:            "webp",
		Usage:           "a tool to encode & decode WebP images",
		UsageText:       "webp <enc|dec> <input> <output>",
		Version:         "<version>",
		HideHelpCommand: true,
		Commands: []*cli.Command{
			{
				Name:      "encode",
				Aliases:   []string{"enc"},
				Usage:     "encode an image to WebP",
				UsageText: "webp enc <input> <output>",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:        "quality",
						Aliases:     []string{"q"},
						Usage:       "image quality between 0-100; higher values result in better quality.",
						Value:       75,
						DefaultText: "75",
						Destination: &quality,
						Required:    false,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					input := command.Args().First()
					output := command.Args().Tail()[0]

					if len(input) == 0 {
						return fmt.Errorf("missing input file")
					}

					if len(output) == 0 {
						return fmt.Errorf("missing output file")
					}

					options := &webp.Options{
						Quality: int(quality),
					}

					now := time.Now()
					img, info, err := encodeWebp(input, output, options)
					duration := time.Since(now)

					if err == nil {
						printResult(img, info, duration, true)
					}

					return err
				},
			},
			{
				Name:      "decode",
				Aliases:   []string{"dec"},
				Usage:     "decode an WebP image to a different format",
				UsageText: "webp dec <input> <output>",
				Action: func(ctx context.Context, command *cli.Command) error {
					input := command.Args().First()
					output := command.Args().Tail()[0]

					if len(input) == 0 {
						return fmt.Errorf("missing input file")
					}

					if len(output) == 0 {
						return fmt.Errorf("missing output file")
					}

					now := time.Now()
					img, info, err := decodeWebp(input, output)
					duration := time.Since(now)

					if err == nil {
						printResult(img, info, duration, false)
					}

					return err
				},
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			return fmt.Errorf("either the command <encode> or <decode> must be used")
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		msg := fmt.Sprintf("🧨 %v", err)
		fmt.Println(red.Render(msg))
	}
}

func printResult(img image.Image, info os.FileInfo, duration time.Duration, isEncode bool) {
	cmd := "decoded"
	if isEncode {
		cmd = "encoded"
	}

	msg := fmt.Sprintf("✅ Successfully %s image to %s in %s",
		cmd, info.Name(), duration.Truncate(time.Millisecond))
	fmt.Println(green.Render(msg))

	msg = fmt.Sprintf("🖼  Image dimensions: %dx%d; size: %d bytes",
		img.Bounds().Dx(), img.Bounds().Dy(), info.Size())
	fmt.Println(yellow.Render(msg))
}
