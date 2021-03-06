package main

import (
	"os"
	"io"
	"fmt"
	"net"
	"flag"
	"sort"
	"bytes"
	"strings"
)

var addr *string = flag.String("address", "127.0.0.1:70", "Specify address to bind to in format '0.0.0.0:70'")
var path *string = flag.String("path", ".", "Specify path to serve, in format '/path/to/serve'")
var host *string = flag.String("host", *addr, "Specify hostname/address to generate links with in format 'xxx.xxx.xxx.xxx:70' (defaults to -address)")

func servefile(c net.Conn, path string) {
	if file, err := os.Open(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
	} else {
		for {
			b := make([]byte, 70)

			if _, err := file.Read(b); err == io.EOF {
				if _, err := c.Write(b); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
					break
				}
				if _, err := io.WriteString(c, ".\r\n"); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
					break
				}
				break
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
				break
			}

			if _, err := c.Write(b); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
				break
			}
		}
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %s\n", err)
		}
	}
}

func filetype(name string) (ftype string) {
	if file, err := os.Stat(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error stating file %s: %s", name, err)
		ftype = "3"
	} else {
		if file.IsDir() {
			ftype = "1"
		} else if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png") {
			ftype = "I"
		} else if strings.HasSuffix(name, ".gif") {
			ftype = "g"
		} else if strings.HasSuffix(name, ".flac") || strings.HasSuffix(name, ".mp3") || strings.HasSuffix(name, ".ogg") || strings.HasSuffix(name, ".wav") {
			ftype = "s"
		} else if file.Mode().IsRegular() {
			ftype = "0"
		} else {
			ftype = "3"
		}
	}
	return ftype
}

func servedir(c net.Conn, path, host, port, reldir string) {
	if dir, err := os.Open(path); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening directory: %s\n", err)
		return
	} else {
		indexfile := fmt.Sprintf("%s/.index", path)
		fmt.Printf("indexfile: %s\n", indexfile)
		if _, err := os.Stat(indexfile); err == nil {
			servefile(c, indexfile)
		} else {
			if names, err := dir.Readdirnames(-1); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading directory: %s\n", err)
				return
			} else {
				sort.Strings(names)
				for i := 0; i < len(names); i++ {
					ftype := filetype(fmt.Sprintf("%s/%s", path, names[i]))
					item := fmt.Sprintf("%s%s\t%s/%s\t%s\t%s\r\n", ftype, names[i], reldir, names[i], host, port)
					if _, err := io.WriteString(c, item); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
					}
				}
			}
			if _, err := io.WriteString(c, ".\r\n"); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
			}
		}
	}
}

func serve(c net.Conn, host string, port string) {
	input := make([]byte, 255)
	if _, err := c.Read(input); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from TCP connection: %s\n", err)
		return
	} else {
		selector := bytes.NewBuffer(input)
		for i := 0; i < len(input); i++ {
			if input[i] == 0 {
				selector.Truncate(i-2)
				break
			}
		}
		file := fmt.Sprintf("./%s", selector.String())
		fmt.Printf("selector: '%s'\n", selector.String())
		fmt.Printf("requested: '%s'\n", file)

		if stat, err := os.Stat(file); err != nil {
			fmt.Fprintf(os.Stderr, "Error stating file: %s\n", err)
		} else {
			if stat.Mode().IsRegular() {
				servefile(c, file)
			} else if stat.IsDir() {
				servedir(c, file, host, port, selector.String())
			} else {
				if _, err := io.WriteString(c, "Error!\r\n.\r\n"); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err)
				}
				if err := c.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err)
				}
			}
		}
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err)
		}
	}
}


func main() {
	flag.Parse()

	if listener, err := net.Listen("tcp", *addr); err != nil {
		fmt.Fprintf(os.Stderr, "Error listening: %s\n", err)
		os.Exit(1)
	} else {
		address := strings.Split(*host, ":")

		if err := os.Chdir(*path); err != nil {
			fmt.Fprintf(os.Stderr, "Error chdiring to %s: %s\n", *path, err)
			os.Exit(1)
		}
		for {
			fmt.Printf("Waiting for new connection...\n")
			if c, err := listener.Accept(); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening TCP connection: %s\n", err)
				os.Exit(1)
			} else {
				fmt.Printf("New connection (%s) with %s\n", c.LocalAddr().String(), c.RemoteAddr().String())
				go serve(c, address[0], address[1])
			}
		}
	}
}
