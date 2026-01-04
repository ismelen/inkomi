package EpubTemplates

import "archive/zip"

const XML = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
	<rootfiles>
	<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
	</rootfiles>
</container>`

const Styles = `@page {
margin: 0;
}
body {
display: block;
margin: 0;
padding: 0;
}
`

var MimeHeader = &zip.FileHeader{
	Name:   "mimetype",
	Method: zip.Store,
}

// HTML
const HTMLStart = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>%s</title>
<link href="../style.css" type="text/css" rel="stylesheet"/>
<meta name="viewport" content="width=%d, height=%d"/>
</head>
<body style="%s">
<div style="text-align:center;top:%.1f%%;">
`

const HTMLImg = `<img width="%d" height="%d" src="%s"/>
</div>
</body>
</html>
`

// NCX
const NCXStart = `<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xml:lang="en-US" xmlns="http://www.daisy.org/z3986/2005/ncx/">
<head>
<meta name="dtb:uid" content="urn:uuid:%s"/>
<meta name="dtb:depth" content="1"/>
<meta name="dtb:totalPageCount" content="0"/>
<meta name="dtb:maxPageNumber" content="0"/>
<meta name="generated" content="true"/>
</head>
<docTitle><text>%s</text></docTitle>
<navMap>`

const NCXNavPoint = `<navPoint id="%s">
<navLabel>
<text>%s</text>
</navLabel>
<content src="%s"/>
</navPoint>
`

const NCXEnd = `</navMap>
</ncx>
`

// NAV
const NAVStart = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>%s</title>
<meta charset="utf-8"/>
</head>
<body>
<nav xmlns:epub="http://www.idpf.org/2007/ops" epub:type="toc" id="toc">
<ol>
`

const NAVLiElem = `<li>
<a href="%s">%s</a>
</li>
`

const NAVBetweenList = `</ol>
</nav>
<nav epub:type="page-list">
<ol>
`

const NAVEnd = `</ol>
</nav>
</body>
</html>
`

// OPF
const OPFStart = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="BookID" 
xmlns="http://www.idpf.org/2007/opf">
<metadata xmlns:opf="http://www.idpf.org/2007/opf" 
xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>%s</dc:title>
<dc:language>en-US</dc:language>
<dc:identifier id="BookID">urn:uuid:%s</dc:identifier>
<dc:contributor id="contributor">EReaderMangaConverter-%s</dc:contributor>
`

const OPFMetas = `<meta property="dcterms:modified">%s</meta>
<meta name="cover" content="cover"/>
<meta name="fixed-layout" content="true"/>
<meta name="original-resolution" content="%dx%d"/>
<meta name="book-type" content="comic"/>
<meta name="primary-writing-mode" content="%s"/>
<meta name="zero-gutter" content="true"/>
<meta name="zero-margin" content="true"/>
<meta name="ke-border-color" content="#FFFFFF"/>
<meta name="ke-border-width" content="0"/>
<meta name="orientation-lock" content="none"/>
<meta name="region-mag" content="%t"/>
<meta property="rendition:spread">landscape</meta>
<meta property="rendition:layout">pre-paginated</meta>
</metadata>

<manifest>
<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
<item id="nav" href="nav.xhtml" properties="nav" media-type="application/xhtml+xml"/>
<item id="cover" href="Images/cover.jpg" media-type="image/jpeg" properties="cover-image"/>
`

const OPFItem = `<item id="%s" href="%s" media-type="%s"/>
`

const OPFPageProgression = `</manifest>
<spine page-progression-direction="%s" toc="ncx">
`

const OPFItemRef = `<itemref idref="page_%s" properties="rendition:page-spread-%s"/>
`

const OPFEnd = `</spine>
</package>
`
