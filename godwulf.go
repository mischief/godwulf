package main

import (
	"os"
	"fmt"
	"net"
	"flag"
	"bytes"
	"strings"
)

var addr *string = flag.String("address", "127.0.0.1:70", "Specify address to bind to in format '0.0.0.0:70'")
var path *string = flag.String("path", ".", "Specify path to serve, in format '/path/to/serve' (NO TRAILING '/')")
var host *string = flag.String("host", *addr, "Specify hostname/address to generate links with in format 'xxx.xxx.xxx.xxx:70'")

func servefile(c net.Conn, path string) {
	if file, err := os.Open(path, os.O_RDONLY, 0); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %s", err.String())
	} else {
		for {
			b := make([]byte, 70)

			if _, err := file.Read(b); err == os.EOF {
				if _, err := c.Write(b); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
					break
				}
				if _, err := c.Write(strings.Bytes(".\r\n")); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
					break
				}
				break
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err.String())
				break
			}

			if _, err := c.Write(b); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
				break
			}
		}
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %s\n", err.String())
		}
	}
}

func filetype(name string) fype string {
	if file, err := os.Stat(name); err != nil {
		fmt.Fprintf(os.Stderr, "Error stating file %s: %s", name, err.String())
		return nil
	} else {
		if file.IsDirectory() {
			return "1"
		} else if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".png") {
			return "I"
		} else if strings.HasSuffix(name, ".gif") {
			return "g"
		} else if file.IsRegular() {
			return "0"
		} else {
			return "3"	// error
		}
	}
}

func servedir(c net.Conn, path, host, port, reldir string) {
	if dir, err := os.Open(path, os.O_RDONLY, 0); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening directory: %s\n", err.String())
		return
	} else {
		indexfile := fmt.Sprintf("%s/.index", path)
		if _, err := os.Stat(indexfile); err == nil {
			servefile(c, indexfile)
		} else {
			if names, err := dir.Readdirnames(-1); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading directory: %s\n", err.String())
				return
			} else {
				for i := 0; i < len(names); i++ {
					ftype := filetype(fmt.Sprintf("%s/%s", path, names[i]))
					item := fmt.Sprintf("%s%s\t%s/%s\t%s\t%s\r\n", ftype, names[i], reldir, names[i], host, port)
					if _, err := c.Write(strings.Bytes(item)); err != nil {
						fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
					}
				}
			}
			if _, err := c.Write(strings.Bytes(".\r\n")); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
			}
		}
	}
}

func serve(c net.Conn, host string, port string) {
	input := make([]byte, 255)
	if _, err := c.Read(input); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from TCP connection: %s\n", err.String())
		return
	} else {
		selector := bytes.NewBuffer(input)
		for i := 0; i < len(input); i++ {
			if input[i] == 0 {
				selector.Truncate(i-2)
				break
			}
		}
		file := fmt.Sprintf("%s%s", *path, selector.String())
		fmt.Printf("selector: '%s'\n", selector.String())
		fmt.Printf("requested: '%s'\n", file)

		if stat, err := os.Stat(file); err != nil {
			fmt.Fprintf(os.Stderr, "Error stating file: %s\n", err.String())
		} else {
			if stat.IsRegular() {
				servefile(c, file)
			} else if stat.IsDirectory() {
				servedir(c, file, host, port, selector.String())
			} else {
				if _, err := c.Write(strings.Bytes("Error!\r\n.\r\n")); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
				}
				if err := c.Close(); err != nil {
					fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err.String())
				}
			}
		}
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err.String())
		}

		/*if selector.String() == "" {
			one := strings.Bytes("iWelcome to godwulf!\r\n")
			two := strings.Bytes(fmt.Sprintf("0About my gopher server\t/about\t%s\t%s\r\n1A directory!\t/dir\t%s\t%s\r\n.\r\n", host, port, host, port))
			msg := bytes.Add(one, two)
			if _, err := c.Write(msg); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
			}
			if err := c.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err.String())
			}
		} else if selector.String() == "/about" {
			msg := strings.Bytes("This is my gopher server, godwulf! It's just in development right now, \r\nso don't worry if stuff doesn't work too well.\r\n.\r\n")
			if _, err := c.Write(msg); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
			}
			if err := c.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err.String())
			}
		} else if selector.String() == "/dir" {
			servedir(requested)
		} else {
			if _, err := c.Write(strings.Bytes("Error!\r\n.\r\n")); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to TCP connection: %s\n", err.String())
			}
			if err := c.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing TCP connection: %s\n", err.String())
			}
		}*/
	}
}


func main() {
	flag.Parse()

	if listener, err := net.Listen("tcp", *addr); err != nil {
		fmt.Fprintf(os.Stderr, "Error listening: %s\n", err.String())
		os.Exit(1)
	} else {
		address := strings.Split(*host, ":", 0)

		if err := os.Chdir(*path); err != nil {
			fmt.Fprintf(os.Stderr, "Error chdiring to %s: %s\n", *path, err.String())
			os.Exit(1)
		}
		for {
			fmt.Printf("Waiting for new connection...\n")
			if c, err := listener.Accept(); err != nil {
				fmt.Fprintf(os.Stderr, "Error opening TCP connection: %s\n", err.String())
				os.Exit(1)
			} else {
				r := c.RemoteAddr()
				l := c.LocalAddr()
				fmt.Printf("New connection (%s) with %s\n", l.String(), r.String())
				go serve(c, address[0], address[1])
			}
		}
	}
}