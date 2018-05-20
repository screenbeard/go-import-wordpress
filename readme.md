# WordPress to Hugo Importer in Golang #

This importer takes a WordPress export XML file and converts each post (through Pandoc) into individual post files in Markdown format ready for use in Hugo.

My own WordPress posts were often written in Textile, so I needed a way to convert from Textile to HTML then HTML to Markdown. This does that conversion then outputs the posts to a file specified from the commandline.

It needs Pandoc installed and available in the path, and needs to be compiled in Golang before running. I'll presume for now that if you are using Hugo that you'll have Go setup and know how to build a program. I imagine you'll also want to alter things to suit your tastes and specific Pandoc conversion needs

To run after compilation:

`go-import-wordpress -filename=files/{wordpress-export.xml} -export="{folder-for-export}"`

Todo:

- ✅ Make a Readme file
- ❎ Add commandline options to tell pandoc the sequence of conversions you wish to perform.
- ❎ Include compiled version for easier download.
