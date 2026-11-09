package epub

const containerXML = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
<rootfiles>
<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
</rootfiles>
</container>`

const styles = `@page {
margin: 0;
}
body {
display: block;
margin: 0;
padding: 0;
}
`

// pageXHTML args: title, viewport width, viewport height, top margin %, img width, img height, img src.
const pageXHTML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>%s</title>
<link href="style.css" type="text/css" rel="stylesheet"/>
<meta name="viewport" content="width=%d, height=%d"/>
</head>
<body style="background-color:#FFFFFF;">
<div style="text-align:center;top:%.1f%%;">
<img width="%d" height="%d" src="%s"/>
</div>
</body>
</html>
`

// ncx args: uuid, title, first page href.
const ncx = `<?xml version="1.0" encoding="UTF-8"?>
<ncx version="2005-1" xml:lang="en-US" xmlns="http://www.daisy.org/z3986/2005/ncx/">
<head>
<meta name="dtb:uid" content="urn:uuid:%[1]s"/>
<meta name="dtb:depth" content="1"/>
<meta name="dtb:totalPageCount" content="0"/>
<meta name="dtb:maxPageNumber" content="0"/>
<meta name="generated" content="true"/>
</head>
<docTitle><text>%[2]s</text></docTitle>
<navMap>
<navPoint id="start">
<navLabel>
<text>%[2]s</text>
</navLabel>
<content src="%[3]s"/>
</navPoint>
</navMap>
</ncx>
`

// nav args: title, first page href.
const nav = `<?xml version="1.0" encoding="utf-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
<title>%[1]s</title>
<meta charset="utf-8"/>
</head>
<body>
<nav xmlns:epub="http://www.idpf.org/2007/ops" epub:type="toc" id="toc">
<ol>
<li><a href="%[2]s">%[1]s</a></li>
</ol>
</nav>
<nav epub:type="page-list">
<ol>
<li><a href="%[2]s">%[1]s</a></li>
</ol>
</nav>
</body>
</html>
`

// opfStart args: title, uuid, modified, width, height, writing mode, cover href, cover media type.
const opfStart = `<?xml version="1.0" encoding="UTF-8"?>
<package version="3.0" unique-identifier="BookID" xmlns="http://www.idpf.org/2007/opf" prefix="rendition: http://www.idpf.org/2013/rendition">
<metadata xmlns:opf="http://www.idpf.org/2007/opf" xmlns:dc="http://purl.org/dc/elements/1.1/">
<dc:title>%s</dc:title>
<dc:language>en-US</dc:language>
<dc:identifier id="BookID">urn:uuid:%s</dc:identifier>
<dc:contributor id="contributor">inkomi</dc:contributor>
<meta property="dcterms:modified">%s</meta>
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
<meta name="region-mag" content="false"/>
<meta property="rendition:spread">landscape</meta>
<meta property="rendition:layout">pre-paginated</meta>
</metadata>
<manifest>
<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
<item id="nav" href="nav.xhtml" properties="nav" media-type="application/xhtml+xml"/>
<item id="css" href="Text/style.css" media-type="text/css"/>
<item id="cover" href="%s" media-type="%s" properties="cover-image"/>
`

// opfItem args: id, href, media type.
const opfItem = `<item id="%s" href="%s" media-type="%s"/>
`

// opfSpineStart args: page progression direction.
const opfSpineStart = `</manifest>
<spine page-progression-direction="%s" toc="ncx">
`

// opfItemRef args: page id, page side.
const opfItemRef = `<itemref idref="%s" properties="rendition:page-spread-%s"/>
`

const opfEnd = `</spine>
</package>
`
