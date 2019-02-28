package downloadgeofabrik

import (
	"strings"
)

type fileFormat struct {
	ext      string
	shortcut rune
	help     string
}

type format struct {
	ID       string `yaml:"ext"`
	Loc      string `yaml:"loc"`
	BasePath string `yaml:"basepath,omitempty"`
	BaseURL  string `yaml:"baseurl,omitempty"`
}

var (
	ffosmBz2 = fileFormat{
		ext:      "osm.bz2",
		help:     "Download osm.bz2 if available",
		shortcut: 'B',
	}
	ffshpZip = fileFormat{
		ext:      "shp.zip",
		help:     "Download shp.zip if available",
		shortcut: 'S',
	}
	ffosmPbf = fileFormat{
		ext:      "osm.pbf",
		help:     "Download osm.pbf (default)",
		shortcut: 'P',
	}
	ffoshPbf = fileFormat{
		ext:      "osh.pbf",
		help:     "Download osh.pbf",
		shortcut: 'H',
	}
	ffstate = fileFormat{
		ext:      "state",
		help:     "Download state.txt file",
		shortcut: 's',
	}
	ffpoly = fileFormat{
		ext:      "poly",
		help:     "Download poly file",
		shortcut: 'p',
	}
	ffkml = fileFormat{
		ext:      "kml",
		help:     "Download kml file",
		shortcut: 'k',
	}
	ffmd5 = fileFormat{ext: "md5"}

	//FileFormats list all file formats supported
	FileFormats = map[string]fileFormat{
		ffosmBz2.ext: ffosmBz2,
		ffshpZip.ext: ffshpZip,
		ffosmPbf.ext: ffosmPbf,
		ffoshPbf.ext: ffoshPbf,
		ffstate.ext:  ffstate,
		ffpoly.ext:   ffpoly,
		ffkml.ext:    ffkml,
	}
	//HashFormats list all hash supported
	HashFormats = map[string]fileFormat{
		ffmd5.ext: ffmd5,
	}
)

//miniFormats get formats of an Element
// and return a string
// according to download-geofabrik short flags.
func miniFormats(s []string) string {
	res := make([]string, 7)
	for _, item := range s {
		switch item {
		case "state":
			res[0] = "s"
		case "osm.pbf":
			res[1] = "P"
		case "osm.bz2":
			res[2] = "B"
		case "osh.pbf":
			res[3] = "H"
		case "poly":
			res[4] = "p"
		case "shp.zip":
			res[5] = "S"
		case "kml":
			res[6] = "k"
		}
	}
	return strings.Join(res, "")
}

func isHashable(c *Config, format string) (bool, string, string) {
	if _, ok := c.Formats[format]; ok {
		for _, h := range HashFormats {
			hash := format + "." + h.ext
			if _, ok := c.Formats[hash]; ok {
				return true, hash, h.ext
			}
		}
	}
	return false, "", ""
}

// getFormats return a pointer to a slice with formats
func getFormats() *[]fileFormat {
	var formatFile []fileFormat
	if *dosmPbf {
		formatFile = append(formatFile, ffosmPbf)
	}
	if *doshPbf {
		formatFile = append(formatFile, ffoshPbf)
	}
	if *dosmBz2 {
		formatFile = append(formatFile, ffosmBz2)
	}
	if *dshpZip {
		formatFile = append(formatFile, ffshpZip)
	}
	if *dstate {
		formatFile = append(formatFile, ffstate)
	}
	if *dpoly {
		formatFile = append(formatFile, ffpoly)
	}
	if *dkml {
		formatFile = append(formatFile, ffkml)
	}
	if len(formatFile) == 0 {
		formatFile = append(formatFile, ffosmPbf)
	}
	return &formatFile
}
