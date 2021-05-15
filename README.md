# mtxconv

A conversion tool to and from Mediocre's MTX format

## Usage

mtxconv can be used to *extract* images from or *bake* them into MTX files. Simply run `mtxconv extract <MTX file>` or `mtxconv bake <image file>` from the commandline and mtxconv will take care of the rest. If you want more control, there are options:

### Global Options

* `--dry-run`: Performs all conversion steps but *doesn't* write the actual output file. Useful for testing without cluttering your storage.
* `--debug`: Enables debug level log messages.

### Options for `mtxconv bake`

* `-q/--jpeg-quality X`: All images you open with mtxconv will be re-encoded as JPEG files. By default, the JPEG quality chosen is 90, which is a good compromise between visual quality and file size. If you want to tweak this value, set this to a number between 0 and 100.
* `-m/--mtx-version X`: mtxconv automatically chooses a suitable MTX version for the image type you supply. Set this to a value between 0 and 2 to override the format.

| Compatibility | MTXv0 | MTXv1 | MTXv2 |
|:--|:--|:--|:--|
| JPEG | ✅ | ✅ | ❌ |
| PNG | ❌ | ✅ | ❌ |
| PVR | ❌ | ❌ | ✅ |

### Options for `mtxconv extract`

There are no special options for the extraction mode.

---

# The MTX Format

Reverse-engineered out of curiosity and documented to the best of my ability

## But First

All the information that follows was gathered during [mtxconv](https://github.com/SamusAranX/mtxconv)'s development. I recommend you just use that instead of trying to build your own tool since mtxconv covers all possible use cases already. This document exists purely to satisfy the nerds.

A few weeks into mtxconv's development, I found out there was a Smash Hit fan wiki containing rudimentary documentation of the MTX format(s), but that documentation was *very* barebones and mostly wrong, so I stuck to my own.

## Here Goes

The games I looked at were:

* [*Smash Hit*](http://www.smashhitgame.com)
* [*Does not Commute*](https://www.mediocre.se/commute/)
* [*Beyondium*](https://www.mediocre.se/beyondium/)
* [*PinOut*](https://www.mediocre.se/pinout/)

It's possible (and likely) that the MTX format was used in other games by Mediocre, but because the listed games were the only games I had purchased, these were the only games I used in my reverse-engineering efforts.

*PinOut* was the last of Mediocre's games and numbering schemes tend to be sequential, so I feel confident in saying that there are only **three** variants of the MTX format, each identified by the first 4 bytes of an MTX file. There's no official naming scheme for these, so I'm just gonna refer to these variants by their header numbers: **MTXv0**, **MTXv1**, and **MTXv2**.

Because v1 and v2 don't meaningfully improve over their "predecessors", I don't think they were meant as "improved" versions of v0 and v1, respectively. Instead, each variant fulfills a specific purpose. More on that below.

For completeness's sake, here's a breakdown of which game uses which format variants:

| Games | MTXv0 | MTXv1 | MTXv2 |
|:--|:--|:--|:--|
| Smash Hit | ✅ | ✅ |  |
| Does not Commute |  | ✅ | ✅ |
| Beyondium | ✅ | ✅ |  | 
| PinOut | ✅ | ✅ | ✅ | 

## General Format Information

Even though these files' names end with `.png.mtx`, there's no actual PNG data in any of them. It's likely that the source files were PNGs and the tool(s) Mediocre used to create MTX files simply appended `.mtx` to their names.

All MTX files use the little endian byte order.

## MTXv0

This format is pretty simple; there's only JPEG data chunks and nothing else. As such, it doesn't support transparency.

The two images contained in (almost) every MTXv0 file are usually the same, except the second image is twice as large in width and height. I assume these are graphical quality tiers rather than mipmaps.
Some MTX files only contain the larger image and omit the first, smaller image by setting the *LengthFirst* field to 0. I further assume those are meant to always be rendered at max quality.

With a hex editor and a trained eye, it's possible to extract JPEG files by hand.

### File Header

| Field | Type | Description |
|--:|:--|:--|
| Magic | uint32 | The magic number. For MTXv0, it's always 0 |
| LengthFirst | uint32 | Amount of bytes in the first image. Can be 0 |
| LengthSecond | uint32 | Amount of bytes in the second image |

### File Structure

| Structure | Length | 
|--:|--:|
| File Header | 12 Bytes |
| First Image | *LengthFirst* Bytes |
| Second Image | *LengthSecond* Bytes |

### Example File

![Crop of the Smash Hit / PinOut vinyl album art](examples/smashhitlp.png)

Crop of the Smash Hit / PinOut vinyl album art

[Download the MTX version](examples/smashhitlp.png.mtx)

## MTXv1

This format introduces alpha masks. Because those are stored as zlib-compressed raw grayscale pixels, extracting the full images by hand is unfeasibly difficult.

Since alpha masks are the only real difference to MTXv0, everything that applied to MTXv0 applies to this format.

When *extracting* images from MTXv1 files, mtxconv will generate PNGs for maximum ease of use. It will however also accept JPEGs when *baking* MTXv1 files, in which case the alpha mask is implied to be opaque for every pixel.

### File Header

| Field | Type | Description |
|--:|:--|:--|
| Magic | uint32 | The magic number. For MTXv1, it's always 1 |
| LengthFirst | uint32 | Amount of bytes in the first image. Can be 0 |
| LengthSecond | uint32 | Amount of bytes in the second image |

This is the same header MTXv0 uses, only the magic number is different.

### Block Header

| Field | Type | Description |
|--:|:--|:--|
| Magic | uint32 | Same as the header, also always 1 |
| Width | uint32 | Width of the following image in pixels |
| Height | uint32 | Ditto, but for the height |

This structure is unique to MTXv1 and precedes every color/mask data block. Width and height are required to reconstruct a mask image from the raw data.

### File Structure

| Structure | Length | 
|--:|--:|
| File Header | 12 Bytes |
| Block Header | 12 Bytes |
| ColorDataLength₁ | 4 Bytes |
| Color Data | *ColorDataLength₁* Bytes |
| MaskDataLength₁ | 4 Bytes |
| Mask Data | *MaskDataLength₁* Bytes |
| Block Header | 12 Bytes |
| ColorDataLength₂ | 4 Bytes |
| Color Data | *ColorDataLength₂* Bytes |
| MaskDataLength₂ | 4 Bytes |
| Mask Data | *MaskDataLength₂* Bytes |

### Example File

![The PinOut logo](examples/pinoutlogo.png)

The PinOut logo, taken from the PinOut press kit

[Download the MTX version](examples/pinoutlogo.png.mtx)

## MTXv2

This format is a thin wrapper around the PVRTC2 format. mtxconv will assist in extracting them from or baking them into MTX files, but it won't convert other image formats to or from PVRTC2. Please use Imagination Technologies's own [PVRTexTool](https://developer.imaginationtech.com/pvrtextool/) for that.

### File Header

| Field | Type | Description |
|--:|:--|:--|
| Magic | uint32 | The magic number. For MTXv2, it's always 2 |
| Unknown | uint16 | Unknown. Is always 256 |

### PVRTC2 Header

Please refer to [this Imagination Technologies document](https://downloads.isee.biz/pub/files/igep-dsp-gst-framework-3_40_00/Graphics_SDK_4_05_00_03/GFX_Linux_SDK/OVG/SDKPackage/Utilities/PVRTexTool/Documentation/PVRTexTool.Reference%20Manual.1.11f.External.pdf), pages 12 and 13.

mtxconv parses this header to determine how much data to extract in order to write a valid PVRTC2 file that PVRTexTool can read.

### Example File

![A generic test card](examples/testcard.png)

A generic "test card" image

[Download the MTX version](examples/testcard.pvr.mtx)