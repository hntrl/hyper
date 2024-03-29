package stdlib

import (
	"fmt"

	"github.com/hntrl/hyper/src/hyper/symbols"
)

type MimeTypesPackage struct{}

func (mt MimeTypesPackage) Get(key string) (symbols.ScopeValue, error) {
	switch key {
	case "MimeType":
		return MimeType, nil
	}
	return nil, nil
}

var (
	MimeType            = MimeTypeClass{}
	MimeTypeDescriptors = &symbols.ClassDescriptors{
		Name: "MimeType",
		Constructors: symbols.ClassConstructorSet{
			symbols.Constructor(symbols.String, func(strValue symbols.StringValue) (MimeTypeValue, error) {
				str := string(strValue)
				for _, v := range types {
					if v == str {
						return MimeTypeValue(v), nil
					}
				}
				return "", fmt.Errorf("unknown mime type %s", str)
			}),
		},
	}
)

type MimeTypeClass struct{}

func (MimeTypeClass) Descriptors() *symbols.ClassDescriptors {
	return MimeTypeDescriptors
}

type MimeTypeValue string

func (MimeTypeValue) Class() symbols.Class {
	return MimeType
}
func (mt MimeTypeValue) Value() interface{} {
	return string(mt)
}

var types = []string{
	"image/aces",
	"image/apng",
	"image/avci",
	"image/avcs",
	"image/avif",
	"image/bmp",
	"image/cgm",
	"image/dicom-rle",
	"image/dpx",
	"image/emf",
	"image/example",
	"image/fits",
	"image/g3fax",
	"image/gif",
	"image/heic",
	"image/heic-sequence",
	"image/heif",
	"image/heif-sequence",
	"image/hej2k",
	"image/hsj2",
	"image/ief",
	"image/jls",
	"image/jp2",
	"image/jpeg",
	"image/jph",
	"image/jphc",
	"image/jpm",
	"image/jpx",
	"image/jxr",
	"image/jxrA",
	"image/jxrS",
	"image/jxs",
	"image/jxsc",
	"image/jxsi",
	"image/jxss",
	"image/ktx",
	"image/ktx2",
	"image/naplps",
	"image/png",
	"image/prs.btif",
	"image/prs.pti",
	"image/pwg-raster",
	"image/svg+xml",
	"image/t38",
	"image/tiff",
	"image/tiff-fx",
	"image/vnd.adobe.photoshop",
	"image/vnd.airzip.accelerator.azv",
	"image/vnd.cns.inf2",
	"image/vnd.dece.graphic",
	"image/vnd.djvu",
	"image/vnd.dwg",
	"image/vnd.dxf",
	"image/vnd.dvb.subtitle",
	"image/vnd.fastbidsheet",
	"image/vnd.fpx",
	"image/vnd.fst",
	"image/vnd.fujixerox.edmics-mmr",
	"image/vnd.fujixerox.edmics-rlc",
	"image/vnd.globalgraphics.pgb",
	"image/vnd.microsoft.icon",
	"image/vnd.mix",
	"image/vnd.ms-modi",
	"image/vnd.mozilla.apng",
	"image/vnd.net-fpx",
	"image/vnd.pco.b16",
	"image/vnd.radiance",
	"image/vnd.sealed.png",
	"image/vnd.sealedmedia.softseal.gif",
	"image/vnd.sealedmedia.softseal.jpg",
	"image/vnd.svf",
	"image/vnd.tencent.tap",
	"image/vnd.valve.source.texture",
	"image/vnd.wap.wbmp",
	"image/vnd.xiff",
	"image/vnd.zbrush.pcx",
	"image/wmf",
	"text/1d-interleaved-parityfec",
	"text/cache-manifest",
	"text/calendar",
	"text/cql",
	"text/cql-expression",
	"text/cql-identifier",
	"text/css",
	"text/csv",
	"text/csv-schema",
	"text/dns",
	"text/encaprtp",
	"text/enriched",
	"text/example",
	"text/fhirpath",
	"text/flexfec",
	"text/fwdred",
	"text/gff3",
	"text/grammar-ref-list",
	"text/hl7v2",
	"text/html",
	"text/javascript",
	"text/jcr-cnd",
	"text/markdown",
	"text/mizar",
	"text/n3",
	"text/parameters",
	"text/parityfec",
	"text/plain",
	"text/provenance-notation",
	"text/prs.fallenstein.rst",
	"text/prs.lines.tag",
	"text/prs.prop.logic",
	"text/raptorfec",
	"text/RED",
	"text/rfc822-headers",
	"text/richtext",
	"text/rtf",
	"text/rtp-enc-aescm128",
	"text/rtploopback",
	"text/rtx",
	"text/SGML",
	"text/shaclc",
	"text/shex",
	"text/spdx",
	"text/strings",
	"text/t140",
	"text/tab-separated-values",
	"text/troff",
	"text/turtle",
	"text/ulpfec",
	"text/uri-list",
	"text/vcard",
	"text/vnd.a",
	"text/vnd.abc",
	"text/vnd.ascii-art",
	"text/vnd.curl",
	"text/vnd.debian.copyright",
	"text/vnd.DMClientScript",
	"text/vnd.dvb.subtitle",
	"text/vnd.esmertec.theme-descriptor",
	"text/vnd.exchangeable",
	"text/vnd.familysearch.gedcom",
	"text/vnd.ficlab.flt",
	"text/vnd.fly",
	"text/vnd.fmi.flexstor",
	"text/vnd.gml",
	"text/vnd.graphviz",
	"text/vnd.hans",
	"text/vnd.hgl",
	"text/vnd.in3d.3dml",
	"text/vnd.in3d.spot",
	"text/vnd.IPTC.NewsML",
	"text/vnd.IPTC.NITF",
	"text/vnd.latex-z",
	"text/vnd.motorola.reflex",
	"text/vnd.ms-mediapackage",
	"text/vnd.net2phone.commcenter.command",
	"text/vnd.radisys.msml-basic-layout",
	"text/vnd.senx.warpscript",
	"text/vnd.si.uricatalogue",
	"text/vnd.sun.j2me.app-descriptor",
	"text/vnd.sosi",
	"text/vnd.trolltech.linguist",
	"text/vnd.wap.si",
	"text/vnd.wap.sl",
	"text/vnd.wap.wml",
	"text/vnd.wap.wmlscript",
	"text/vtt",
	"text/xml",
	"text/xml-external-parsed-entity",
	"video/1d-interleaved-parityfec",
	"video/3gpp",
	"video/3gpp2",
	"video/3gpp-tt",
	"video/AV1",
	"video/BMPEG",
	"video/BT656",
	"video/CelB",
	"video/DV",
	"video/encaprtp",
	"video/example",
	"video/FFV1",
	"video/flexfec",
	"video/H261",
	"video/H263",
	"video/H263-1998",
	"video/H263-2000",
	"video/H264",
	"video/H264-RCDO",
	"video/H264-SVC",
	"video/H265",
	"video/H266",
	"video/iso.segment",
	"video/JPEG",
	"video/jpeg2000",
	"video/jxsv",
	"video/mj2",
	"video/MP1S",
	"video/MP2P",
	"video/MP2T",
	"video/mp4",
	"video/MP4V-ES",
	"video/MPV",
	"video/mpeg",
	"video/mpeg4-generic",
	"video/nv",
	"video/ogg",
	"video/parityfec",
	"video/pointer",
	"video/quicktime",
	"video/raptorfec",
	"video/raw",
	"video/rtp-enc-aescm128",
	"video/rtploopback",
	"video/rtx",
	"video/scip",
	"video/smpte291",
	"video/SMPTE292M",
	"video/ulpfec",
	"video/vc1",
	"video/vc2",
	"video/vnd.CCTV",
	"video/vnd.dece.hd",
	"video/vnd.dece.mobile",
	"video/vnd.dece.mp4",
	"video/vnd.dece.pd",
	"video/vnd.dece.sd",
	"video/vnd.dece.video",
	"video/vnd.directv.mpeg",
	"video/vnd.directv.mpeg-tts",
	"video/vnd.dlna.mpeg-tts",
	"video/vnd.dvb.file",
	"video/vnd.fvt",
	"video/vnd.hns.video",
	"video/vnd.iptvforum.1dparityfec-1010",
	"video/vnd.iptvforum.1dparityfec-2005",
	"video/vnd.iptvforum.2dparityfec-1010",
	"video/vnd.iptvforum.2dparityfec-2005",
	"video/vnd.iptvforum.ttsavc",
	"video/vnd.iptvforum.ttsmpeg2",
	"video/vnd.motorola.video",
	"video/vnd.motorola.videop",
	"video/vnd.mpegurl",
	"video/vnd.ms-playready.media.pyv",
	"video/vnd.nokia.interleaved-multimedia",
	"video/vnd.nokia.mp4vr",
	"video/vnd.nokia.videovoip",
	"video/vnd.objectvideo",
	"video/vnd.radgamettools.bink",
	"video/vnd.radgamettools.smacker",
	"video/vnd.sealed.mpeg1",
	"video/vnd.sealed.mpeg4",
	"video/vnd.sealed.swf",
	"video/vnd.sealedmedia.softseal.mov",
	"video/vnd.uvvu.mp4",
	"video/vnd.youtube.yt",
	"video/vnd.vivo",
	"video/VP8",
	"video/VP9",
	"font/collection",
	"font/otf",
	"font/sfnt",
	"font/ttf",
	"font/woff",
	"font/woff2",
	"audio/1d-interleaved-parityfec",
	"audio/32kadpcm",
	"audio/3gpp",
	"audio/3gpp2",
	"audio/aac",
	"audio/ac3",
	"audio/AMR",
	"audio/AMR-WB",
	"audio/amr-wb+",
	"audio/aptx",
	"audio/asc",
	"audio/ATRAC-ADVANCED-LOSSLESS",
	"audio/ATRAC-X",
	"audio/ATRAC3",
	"audio/basic",
	"audio/BV16",
	"audio/BV32",
	"audio/clearmode",
	"audio/CN",
	"audio/DAT12",
	"audio/dls",
	"audio/dsr-es201108",
	"audio/dsr-es202050",
	"audio/dsr-es202211",
	"audio/dsr-es202212",
	"audio/DV",
	"audio/DVI4",
	"audio/eac3",
	"audio/encaprtp",
	"audio/EVRC",
	"audio/EVRC-QCP",
	"audio/EVRC0",
	"audio/EVRC1",
	"audio/EVRCB",
	"audio/EVRCB0",
	"audio/EVRCB1",
	"audio/EVRCNW",
	"audio/EVRCNW0",
	"audio/EVRCNW1",
	"audio/EVRCWB",
	"audio/EVRCWB0",
	"audio/EVRCWB1",
	"audio/EVS",
	"audio/example",
	"audio/flexfec",
	"audio/fwdred",
	"audio/G711-0",
	"audio/G719",
	"audio/G7221",
	"audio/G722",
	"audio/G723",
	"audio/G726-16",
	"audio/G726-24",
	"audio/G726-32",
	"audio/G726-40",
	"audio/G728",
	"audio/G729",
	"audio/G7291",
	"audio/G729D",
	"audio/G729E",
	"audio/GSM",
	"audio/GSM-EFR",
	"audio/GSM-HR-08",
	"audio/iLBC",
	"audio/ip-mr_v2.5",
	"audio/L8",
	"audio/L16",
	"audio/L20",
	"audio/L24",
	"audio/LPC",
	"audio/MELP",
	"audio/MELP600",
	"audio/MELP1200",
	"audio/MELP2400",
	"audio/mhas",
	"audio/mobile-xmf",
	"audio/MPA",
	"audio/mp4",
	"audio/MP4A-LATM",
	"audio/mpa-robust",
	"audio/mpeg",
	"audio/mpeg4-generic",
	"audio/ogg",
	"audio/opus",
	"audio/parityfec",
	"audio/PCMA",
	"audio/PCMA-WB",
	"audio/PCMU",
	"audio/PCMU-WB",
	"audio/prs.sid",
	"audio/QCELP",
	"audio/raptorfec",
	"audio/RED",
	"audio/rtp-enc-aescm128",
	"audio/rtploopback",
	"audio/rtp-midi",
	"audio/rtx",
	"audio/scip",
	"audio/SMV",
	"audio/SMV0",
	"audio/SMV-QCP",
	"audio/sofa",
	"audio/sp-midi",
	"audio/speex",
	"audio/t140c",
	"audio/t38",
	"audio/telephone-event",
	"audio/TETRA_ACELP",
	"audio/TETRA_ACELP_BB",
	"audio/tone",
	"audio/TSVCIS",
	"audio/UEMCLIP",
	"audio/ulpfec",
	"audio/usac",
	"audio/VDVI",
	"audio/VMR-WB",
	"audio/vnd.3gpp.iufp",
	"audio/vnd.4SB",
	"audio/vnd.audiokoz",
	"audio/vnd.CELP",
	"audio/vnd.cisco.nse",
	"audio/vnd.cmles.radio-events",
	"audio/vnd.cns.anp1",
	"audio/vnd.cns.inf1",
	"audio/vnd.dece.audio",
	"audio/vnd.digital-winds",
	"audio/vnd.dlna.adts",
	"audio/vnd.dolby.heaac.1",
	"audio/vnd.dolby.heaac.2",
	"audio/vnd.dolby.mlp",
	"audio/vnd.dolby.mps",
	"audio/vnd.dolby.pl2",
	"audio/vnd.dolby.pl2x",
	"audio/vnd.dolby.pl2z",
	"audio/vnd.dolby.pulse.1",
	"audio/vnd.dra",
	"audio/vnd.dts",
	"audio/vnd.dts.hd",
	"audio/vnd.dts.uhd",
	"audio/vnd.dvb.file",
	"audio/vnd.everad.plj",
	"audio/vnd.hns.audio",
	"audio/vnd.lucent.voice",
	"audio/vnd.ms-playready.media.pya",
	"audio/vnd.nokia.mobile-xmf",
	"audio/vnd.nortel.vbk",
	"audio/vnd.nuera.ecelp4800",
	"audio/vnd.nuera.ecelp7470",
	"audio/vnd.nuera.ecelp9600",
	"audio/vnd.octel.sbc",
	"audio/vnd.presonus.multitrack",
	"audio/vnd.qcelp",
	"audio/vnd.rhetorex.32kadpcm",
	"audio/vnd.rip",
	"audio/vnd.sealedmedia.softseal.mpeg",
	"audio/vnd.vmx.cvsd",
	"audio/vorbis",
	"audio/vorbis-config",
}
