# webp-go

A Go library and CLI tool to encode/decode WebP images without system dependencies (CGO).

## 💡 Motivation

There are a couple of libraries to encode WebP images in Go, and even though they do the job well, they have one limitation that don't satisfy my needs: they either depend on libraries to be installed on the system to be built and/or later be executed.

**webp-go** uses CGO to create a static implementation of WebP, so you don't need to have `libwebp` (or any of its sub-dependencies) installed to encode WebP images.

It also runs on native code (supports `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`), so it achieves the best performance possible.

## ⬇️ Installation

### Library

This library can be installed using Go modules. To do that, run the following command in your project's root directory:

```bash
$ go get github.com/vegidio/webp-go
```

### CLI

The binaries are available for Windows, macOS, and Linux. Download the [latest release](https://github.com/vegidio/webp-go/releases) that matches your computer architecture and operating system.

## 🤖 Usage

### Library

This is a CGO library, so to use it, you _must_ enable CGO while building your application. You can do that by setting the `CGO_ENABLED` environment variable to `1`:

```bash
$ CGO_ENABLED=1 go build /path/to/your/app.go
```

Here are some examples of how to encode and decode HEIC images using this library. These snippets don't have any error handling for the sake of simplicity, but you should always check for errors in production code.

#### Encoding

```go
var originalImage image.Image = ... // an image.Image to be encoded
webpFile, err := os.Create("/path/to/image.webp") // create the file to save the WebP
err = webp.Encode(webpFile, originalImage, nil) // encode the image & save it to the file
```

#### Decoding

```go
import _ "github.com/vegidio/webp-go" // do a blank import to register the HEIC decoder
webpFile, err := os.Open("/path/to/image.webp") // open the WebP file to be decoded
webpImage, _, err := image.Decode(webpFile) // decode the image
```

### CLI

If you want to decode a WebP image, run the following command:

```bash
$ webp decode /path/to/image.webp /path/to/image.png
```

---

To encode an image to WebP, run the following command:

```bash
$ webp encode /path/to/image.png /path/to/image.webp
```

For the full list of parameters, type `webp encode --help` in the terminal.

## 💣 Troubleshooting

### I cannot build my app after importing this library

If you cannot build your app after importing **webp-go**, it is probably because you didn't set the `CGO_ENABLED` environment variable to `1`.

You must either set a global environment variable with `export CGO_ENABLED=1` or set it in the command line when building your app with `CGO_ENABLED=1 go build /path/to/your/app.go`.

### "App Is Damaged/Blocked..." (Windows & macOS only)

For a couple of years now, Microsoft and Apple have required developers to join their "Developer Program" to gain the pretentious status of an _identified developer_ 😛.

Translating to non-BS language, this means that if you’re not registered with them (i.e., paying the fee), you can’t freely distribute Windows or macOS software. Apps from unidentified developers will display a message saying the app is damaged or blocked and can’t be opened.

To bypass this, open the Terminal and run one of the commands below (depending on your operating system), replacing `<path-to-app>` with the correct path to where you’ve installed the app:

- Windows: `Unblock-File -Path <path-to-app>`
- macOS: `xattr -d com.apple.quarantine <path-to-app>`

## 📝 License

**webp-go** is released under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

## 👨🏾‍💻 Author

Vinicius Egidio ([vinicius.io](http://vinicius.io))
