package mtx

import (
	"bytes"
	"encoding/binary"
	"os"
)

/*
The following header sizes are Known™ so I'm hardcoding them
A dynamic (instead of constant) way to obtain struct sizes would be
int(reflect.TypeOf(HeaderV0{}).Size())
*/
const (
	HEADER_V0_SIZE = 12
	HEADER_V1_SIZE = 12
	HEADER_V2_SIZE = 6

	BLOCK_HEADER_V1_SIZE = 12

	PVRTC2_HEADER_SIZE = 52

	MAX_IMAGE_SIZE      = 4096
	MAX_INPUT_FILE_SIZE = 1073741824 // 1 GiB
)

// HeaderV0 represents an MTX v0 header, used in Smash Hit, Beyondium, and PinOut
type HeaderV0 struct {
	Magic             uint32
	FirstImageLength  uint32
	SecondImageLength uint32
}

// HeaderV1 represents an MTX v1 header, used in Smash Hit, Beyondium, Does Not Commute, and PinOut
type HeaderV1 struct {
	Magic             uint32
	SecondBlockOffset uint32

	/*
		this is probably something different,
		but I'm calling it SizeCheck because it
		doesn't seem to have another purpose
	*/
	SizeCheck uint32
}

// BlockHeaderV1 represents the header structure of image/mask blocks in MTXv1 files
type BlockHeaderV1 struct {
	Magic  uint32
	Width  uint32
	Height uint32
}

// HeaderV2 represents an MTX v2 header, used in Does Not Commute and PinOut
type HeaderV2 struct {
	Magic   uint32
	Unknown uint16 // appears to always be 256
}

// PVRTC2Header represents the header of a PVRTC2 file. See also https://downloads.isee.biz/pub/files/igep-dsp-gst-framework-3_40_00/Graphics_SDK_4_05_00_03/GFX_Linux_SDK/OVG/SDKPackage/Utilities/PVRTexTool/Documentation/PVRTexTool.Reference%20Manual.1.11f.External.pdf
type PVRTC2Header struct {
	HeaderSize         uint32
	Height             uint32
	Width              uint32
	MipMapCount        uint32
	PixelFormatFlags   uint32
	CompressedDataSize uint32
	BitCount           uint32
	BitMaskR           uint32
	BitMaskG           uint32
	BitMaskB           uint32
	BitMaskA           uint32
	Magic              [4]byte
	NumSurfaces        uint32
}

func readHeaderV0(file *os.File) (HeaderV0, error) {
	header := HeaderV0{}

	if headerData, err := readSomeBytes(file, HEADER_V0_SIZE); err != nil {
		return header, err
	} else {
		headerBuf := bytes.NewBuffer(headerData)
		err := binary.Read(headerBuf, binary.LittleEndian, &header)
		if err != nil {
			return header, err
		}
	}

	return header, nil
}

func readHeaderV1(file *os.File) (HeaderV1, error) {
	header := HeaderV1{}

	if headerData, err := readSomeBytes(file, HEADER_V1_SIZE); err != nil {
		return header, err
	} else {
		headerBuf := bytes.NewBuffer(headerData)
		err := binary.Read(headerBuf, binary.LittleEndian, &header)
		if err != nil {
			return header, err
		}
	}

	return header, nil
}

func readBlockHeaderV1(file *os.File) (BlockHeaderV1, error) {
	header := BlockHeaderV1{}

	if headerData, err := readSomeBytes(file, BLOCK_HEADER_V1_SIZE); err != nil {
		return header, err
	} else {
		headerBuf := bytes.NewBuffer(headerData)
		err := binary.Read(headerBuf, binary.LittleEndian, &header)
		if err != nil {
			return header, err
		}
	}

	return header, nil
}

func readHeaderV2(file *os.File) (HeaderV2, error) {
	header := HeaderV2{}

	if headerData, err := readSomeBytes(file, HEADER_V2_SIZE); err != nil {
		return header, err
	} else {
		headerBuf := bytes.NewBuffer(headerData)
		err := binary.Read(headerBuf, binary.LittleEndian, &header)
		if err != nil {
			return header, err
		}
	}

	return header, nil
}

func readPVRTC2Header(file *os.File) (PVRTC2Header, error) {
	header := PVRTC2Header{}

	if headerData, err := readSomeBytes(file, PVRTC2_HEADER_SIZE); err != nil {
		return header, err
	} else {
		headerBuf := bytes.NewBuffer(headerData)
		err := binary.Read(headerBuf, binary.LittleEndian, &header)
		if err != nil {
			return header, err
		}
	}

	return header, nil
}
