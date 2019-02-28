// package downloadgeofabrik is a tool to download a lot of things

package downloadgeofabrik

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "2.3.0"

var (
	app         = kingpin.New("download-geofabrik", "A command-line tool for downloading OSM files.")
	fService    = app.Flag("service", "Can switch to another service. You can use \"geofabrik\", \"openstreetmap.fr\" or \"gislab\". It automatically change config file if -c is unused.").Default("geofabrik").String()
	fConfig     = app.Flag("config", "Set Config file.").Default("./geofabrik.yml").Short('c').String()
	fNodownload = app.Flag("nodownload", "Do not download file (test only)").Short('n').Bool()
	fVerbose    = app.Flag("verbose", "Be verbose").Short('v').Bool()
	fQuiet      = app.Flag("quiet", "Be quiet").Short('q').Bool()
	fProgress   = app.Flag("progress", "Add a progress bar").Bool()
	fProxyHTTP  = app.Flag("proxy-http", "Use http proxy, format: proxy_address:port").Default("").String()
	fProxySock5 = app.Flag("proxy-sock5", "Use Sock5 proxy, format: proxy_address:port").Default("").String()
	fProxyUser  = app.Flag("proxy-user", "Proxy user").Default("").String()
	fProxyPass  = app.Flag("proxy-pass", "Proxy password").Default("").String()

	update = app.Command("update", "Update geofabrik.yml from github *** DEPRECATED you should prefer use generate ***")
	fURL   = update.Flag("url", "Url for config source").Default("https://raw.githubusercontent.com/julien-noblet/download-geofabrik/master/geofabrik.yml").String()

	list = app.Command("list", "Show elements available")
	lmd  = list.Flag("markdown", "generate list in Markdown format").Bool()

	download = app.Command("download", "Download element") //TODO : add d as command
	delement = download.Arg("element", "OSM element").Required().String()
	dosmBz2  = download.Flag(ffosmBz2.ext, ffosmBz2.help).Short(ffosmBz2.shortcut).Bool()
	dshpZip  = download.Flag(ffshpZip.ext, ffshpZip.help).Short(ffshpZip.shortcut).Bool()
	dosmPbf  = download.Flag(ffosmPbf.ext, ffosmPbf.help).Short(ffosmPbf.shortcut).Bool()
	doshPbf  = download.Flag(ffoshPbf.ext, ffoshPbf.help).Short(ffoshPbf.shortcut).Bool()
	dstate   = download.Flag(ffstate.ext, ffstate.help).Short(ffstate.shortcut).Bool()
	dpoly    = download.Flag(ffpoly.ext, ffpoly.help).Short(ffpoly.shortcut).Bool()
	dkml     = download.Flag(ffkml.ext, ffkml.help).Short(ffkml.shortcut).Bool()
	dCheck   = download.Flag("check", "Control with checksum (default) Use --no-check to discard control").Default("true").Bool()

	generate = app.Command("generate", "Generate a new config file")
)

func listAllRegions(c Config, format string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader([]string{"ShortName", "Is in", "Long Name", "formats"})
	if format == "Markdown" {
		table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
		table.SetCenterSeparator("|")
	}
	keys := make(sort.StringSlice, len(c.Elements))
	i := 0
	for k := range c.Elements {
		keys[i] = k
		i++
	}
	keys.Sort()
	for _, item := range keys {
		table.Append([]string{item, c.Elements[c.Elements[item].Parent].Name, c.Elements[item].Name, miniFormats(c.Elements[item].Formats)})
	}
	table.Render()
	fmt.Printf("Total elements: %#v\n", len(c.Elements))
}

// UpdateConfig : simple script to download lastest config from repo
func UpdateConfig(myURL string, myconfig string) error {
	if !*fQuiet {
		log.Print("*** DEPRECATED you should prefer use generate ***")
	}
	err := downloadFromURL(myURL, myconfig)
	if err != nil {
		if *fVerbose {
			log.Println(err)
		}
		return fmt.Errorf("Can't updating %v please use generate", myconfig)
	}
	if *fVerbose && !*fQuiet {
		log.Println("Congratulation, you have the latest geofabrik.yml")
	}
	return nil
}

func checkService() bool {
	switch *fService {
	case "geofabrik":
		return true
	case "openstreetmap.fr":
		if strings.EqualFold(*fConfig, "./geofabrik.yml") {
			*fConfig = "./openstreetmap.fr.yml"
		}
		return true
	case "gislab":
		if strings.EqualFold(*fConfig, "./geofabrik.yml") {
			*fConfig = "./gislab.yml"
		}
		return true
	}
	return false
}

func catch(err error) {
	if err != nil {
		log.Fatalln(err.Error()) // Fatalln is better than Panic or Println
		// Println only log but dont do exit(1),
		// Panic add a lot of verbose detail for debug but it's too aggressive!
	}
}

func listCommand() {
	var format = ""
	if *lmd {
		format = "Markdown"
	}
	configPtr, err := LoadConfig(*fConfig)
	catch(err)
	listAllRegions(*configPtr, format)
}

func downloadCommand() {
	configPtr, err := LoadConfig(*fConfig)
	catch(err)
	formatFile := getFormats()
	for _, format := range *formatFile {
		if ok, _, _ := isHashable(configPtr, format.ext); *dCheck && ok {
			if IsFileExist(*delement + "." + format.ext) {
				if !downloadChecksum(format.ext) {
					if !*fQuiet {
						log.Println("Checksum mismatch, re-downloading", *delement+"."+format.ext)
					}
					myElem, err := findElem(configPtr, *delement)
					catch(err)
					myURL, err := Element2URL(configPtr, myElem, format.ext)
					catch(err)
					err = downloadFromURL(myURL, *delement+"."+format.ext)
					catch(err)
					downloadChecksum(format.ext)
				} else {
					if !*fQuiet {
						log.Printf("Checksum match, no download!")
					}
				}
			} else {
				myElem, err := findElem(configPtr, *delement)
				catch(err)
				myURL, err := Element2URL(configPtr, myElem, format.ext)
				catch(err)
				err = downloadFromURL(myURL, *delement+"."+format.ext)
				catch(err)
				if !downloadChecksum(format.ext) && !*fQuiet {
					log.Println("Checksum mismatch, please re-download", *delement+"."+format.ext)
				}
			}
		} else {
			myElem, err := findElem(configPtr, *delement)
			catch(err)
			myURL, err := Element2URL(configPtr, myElem, format.ext)
			catch(err)
			err = downloadFromURL(myURL, *delement+"."+format.ext)
			catch(err)
		}
	}
}

// Main function, load app and grab commands
func Main() {
	app.Version(version) // Add version flag
	commands := kingpin.MustParse(app.Parse(os.Args[1:]))
	checkService()
	switch commands {
	case list.FullCommand():
		listCommand()
	case update.FullCommand():
		err := UpdateConfig(*fURL, *fConfig)
		catch(err)
	case download.FullCommand():
		downloadCommand()
	case generate.FullCommand():
		Generate(*fConfig)
	}
}

// IsFileExist check if filePath is valid and exist or not
func IsFileExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}

//HashFileMD5 open filePath and try to md5sum it.
func HashFileMD5(filePath string) (string, error) {
	var returnMD5String string
	if IsFileExist(filePath) {
		file, err := os.Open(filePath)
		if err != nil {
			return returnMD5String, err
		}
		defer func() {
			err := file.Close()
			catch(err)
		}()
		hash := md5.New()

		if _, err := io.Copy(hash, file); err != nil {
			return returnMD5String, err
		}
		hashInBytes := hash.Sum(nil)[:16]
		returnMD5String = hex.EncodeToString(hashInBytes)
		return returnMD5String, nil
	}
	return returnMD5String, nil
}

func controlHash(hashfile string, hash string) (bool, error) {
	if IsFileExist(hashfile) {
		file, err := ioutil.ReadFile(hashfile)
		if err != nil {
			return false, err
		}
		filehash := strings.Split(string(file), " ")[0]
		if *fVerbose && !*fQuiet {
			log.Println("Hash from file :", filehash)
		}
		return strings.EqualFold(hash, filehash), nil
	}
	return false, nil
}

func downloadChecksum(format string) bool {
	// TODO: use HashFormats!!!!
	ret := false
	if *dCheck {
		fhash := format + "." + ffmd5.ext
		configPtr, err := LoadConfig(*fConfig)
		catch(err)
		if ok, _, _ := isHashable(configPtr, format); ok {
			myElem, err := findElem(configPtr, *delement)
			catch(err)
			myURL, err := Element2URL(configPtr, myElem, fhash)
			catch(err)
			err = downloadFromURL(myURL, *delement+"."+fhash)
			catch(err)
			if *fVerbose && !*fQuiet {
				log.Println("Hashing", *delement+"."+format)
			}
			hashed, err := HashFileMD5(*delement + "." + format)
			if err != nil {
				log.Panic(fmt.Errorf(err.Error()))
			}
			if *fVerbose && !*fQuiet {
				log.Println("MD5 :", hashed)
			}
			ret, err := controlHash(*delement+"."+fhash, hashed)
			if err != nil {
				log.Panic(fmt.Errorf(err.Error()))
			}
			if !*fQuiet {
				if ret {
					log.Println("Checksum OK for", *delement+"."+format)
				} else {
					log.Println("Checksum MISMATCH for", *delement+"."+format)
				}
			}
			return ret
		}
		if !*fQuiet {
			log.Println("No checksum provided for", *delement+"."+format)
		}
	}
	return ret
}
