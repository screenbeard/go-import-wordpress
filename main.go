package main

import (
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "time"
    "net/url"
    "encoding/xml"
    "encoding/json"
    "os"
    "os/exec"
    "bytes"
)

var filename string
var export string

func init() {
    flag.StringVar(&filename, "filename", "", "name of the file you wish to import")
    flag.StringVar(&export, "export", "export", "location of the directory you want to export to")
    flag.Parse()
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}


func elapsed(what string) func() {
    start := time.Now()
    return func() {
        fmt.Printf("%s took %v\n", what, time.Since(start))
    }
}

func main() {

    defer elapsed("script")()

    if filename != "" {
        xmlFile, err := os.Open(filename)
        check(err)

        fmt.Println("Successfully Opened " + filename)
        defer xmlFile.Close()

        byteValue, _ := ioutil.ReadAll(xmlFile)
        var rss RSS
        xml.Unmarshal(byteValue, &rss)

        for i := 0; i < len(rss.Channels); i++ {
            // fmt.Println("Channel title: " + rss.Channels[i].Title)
            // fmt.Println("Channel link: " + rss.Channels[i].Link)
            // fmt.Println("Channel description: " + rss.Channels[i].Description)
            // fmt.Println("Channel language: " + rss.Channels[i].Language)
            // fmt.Println("Channel URL: " + rss.Channels[i].BaseSiteURL)
            // fmt.Println("Channel Publish date: " + rss.Channels[i].PubDate)

            for j := 0; j < len(rss.Channels[i].Items); j++ {
                if rss.Channels[i].Items[j].PostType == "post" {
                    var post Post

                    post.postType = "post"
                    if rss.Channels[i].Items[j].Status == "draft" {
                        post.frontmatter.Draft = true
                    } else {
                        post.frontmatter.setPubDate(rss.Channels[i].Items[j].PublishDate)
                    }
                    post.setContent(rss.Channels[i].Items[j].Content, "--from=textile", "--to=html")
                    post.setContent(rss.Channels[i].Items[j].Content, "--from=html", "--to=markdown")
                    post.frontmatter.setSlug(rss.Channels[i].Items[j].Slug)
                    post.frontmatter.Date = rss.Channels[i].Items[j].Date
                    check(err)

                    for k := 0; k < len(rss.Channels[i].Items[j].Categories); k++ {
                        post.frontmatter.addTaxonomy(rss.Channels[i].Items[j].Categories[k].Domain, rss.Channels[i].Items[j].Categories[k].Name)
                    }
                    post.Write()
                }
            }
        }
    } else {
        println("No filename included")
    }
}

type RSS struct {
    XMLName xml.Name `xml:"rss"`
    Channels []Channel `xml:"channel"`
}

type Channel struct {
    XMLName xml.Name `xml:"channel"`
    Title string `xml:"title"`
    Link string `xml:"link"`
    Description string `xml:"description"`
    PubDate string `xml:"pubDate"`
    Language string `xml:"language"`
    BaseSiteURL string `xml:"base_site_url"`
    BaseBlogURL string `xml:"base_blog_url"`
    Items []Item `xml:"item"`
}

type Item struct {
    XMLName xml.Name `xml:"item"`
    Aliases []string
    PostType string `xml:"http://wordpress.org/export/1.2/ post_type"`
    Date string `xml:"http://wordpress.org/export/1.2/ post_date_gmt"`
    Description string `xml:"Description"`
    ExpiryDate string
    LinkTitle string `xml:"title"`
    PublishDate string `xml:"pubDate"`
    Slug string `xml:"http://wordpress.org/export/1.2/ post_name"`
    Status string `xml:"http://wordpress.org/export/1.2/ status"`
    Url string `xml:"link"`
    Content string `xml:"http://purl.org/rss/1.0/modules/content/ encoded"`
    Categories []Category `xml:"category"`
}

type Category struct {
    Domain string `xml:"domain,attr"`
    NiceName string `xml:"nicename,attr"`
    Name string `xml:",cdata"`
}

type Post struct {
    postType string
    frontmatter FrontMatter
    content string
}

func (p Post) Write(){
    f, err := os.Create(export + "/" + p.frontmatter.Slug + ".md")
    check(err)
    defer f.Close()

    b, err := json.MarshalIndent(p.frontmatter, "", "    ")
    _, err = f.Write(b)
    _, err = f.WriteString("\n\n" + p.content + "\n")
    check(err)
    f.Sync()
}

func (p *Post) setContent(content string, convertfrom string, convertto string) string {

    pandocExec, err := exec.LookPath("pandoc")
    check(err)

    cmdArgs := []string{convertfrom, convertto}

    cmd := exec.Command(pandocExec, cmdArgs...)

    stdin, err := cmd.StdinPipe()
    check(err)

    stdout, err := cmd.StdoutPipe()
    err = cmd.Start();
    check(err)

    io.WriteString(stdin, content)
    stdin.Close()

    buf := new(bytes.Buffer)
    buf.ReadFrom(stdout)
    p.content = buf.String()
    cmd.Wait()

    return p.content
}

type FrontMatter struct {
    Slug string `json:"slug"`
    Title string `json:"title"`
    Draft bool `json:"draft,omitempty"`
    Date string `json:"date"`
    PublishDate string `json:"publishDate"`
    Url string `json:"url"`
    Tags []string `json:"tags"`
    Categories []string `json:"categories"`
}

func (f *FrontMatter) setSlug(slug string) string {
    tmp, err := url.Parse(slug)
    check(err)

    if len(tmp.Path) > 100 {
        tmp.Path = tmp.Path[:99]
    }

    f.Slug = tmp.Path
    return f.Slug
}

func (f *FrontMatter) setPubDate(date string) string {
    tmp, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", date)
    check(err)

    f.PublishDate = tmp.Format("2006-01-02 15:04:05")
    return f.PublishDate
}

func (f *FrontMatter) addTaxonomy(taxonomyType string, name string) {
    if taxonomyType == "category" {
        f.Categories = append(f.Categories, name)
    } else if taxonomyType == "tag" {
        f.Tags = append(f.Tags, name)
    }
}
