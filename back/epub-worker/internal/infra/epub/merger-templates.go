package epub

const MergerContainerXML = `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
`

const MergerOPFStart = `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="pub-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:identifier id="pub-id">urn:uuid:%s</dc:identifier>
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>es</dc:language>
    <meta property="dcterms:modified">%s</meta>
    <meta name="cover" content="%s"/>
  </metadata>
  <manifest>
`

const MergerOPFItem = `    <item id="%s" href="%s" media-type="%s"%s/>
`

const MergerOPFSpineStart = `  </manifest>
  <spine toc="%s">
`

const MergerOPFItemRef = `    <itemref idref="%s"/>
`

const MergerOPFEnd = `  </spine>
</package>
`

const MergerNavStart = `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head><title>%s</title></head>
<body>
  <nav epub:type="toc" id="toc">
    <h1>%s</h1>
    <ol>
`

const MergerNavEntry = `      <li><a href="%s">%s</a></li>
`

const MergerNavEnd = `    </ol>
  </nav>
</body>
</html>
`

const MergerNCXStart = `<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="urn:uuid:%s"/>
  </head>
  <docTitle><text>%s</text></docTitle>
  <navMap>
`

const MergerNCXEntry = `    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel><text>%s</text></navLabel>
      <content src="%s"/>
    </navPoint>
`

const MergerNCXEnd = `  </navMap>
</ncx>
`
