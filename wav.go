package wav

import (
	"bytes"
	"encoding/binary"
	"io"
)

// golang types
// uint8       the set of all unsigned  8-bit (1 byte) integers (0 to 255)
// uint16      the set of all unsigned 16-bit (2 byte) integers (0 to 65535)
// uint32      the set of all unsigned 32-bit (4 byte) integers (0 to 4294967295)
// uint64      the set of all unsigned 64-bit (8 byte) integers (0 to 18446744073709551615)

// http://soundfile.sapp.org/doc/WaveFormat/
// The default byte ordering assumed for WAVE data files is little-endian.
// Files written using the big-endian byte ordering scheme have the identifier RIFX instead of RIFF.
type wavfileHeader struct {

	// Offset 0
	// Contains the letters "RIFF" in ASCII form
	riffTag [4]uint8 // ChunkID 4 bytes

	// Offset 4
	// This is the size of the entire file in bytes minus 8 bytes for the two fields not included in this count: ChunkID and ChunkSize.
	// Also = 4 + (8 + SubChunk1Size) + (8 + SubChunk2Size)
	riffLength uint32 // ChunkSize 4 bytes

	// Offset 8
	// Contains the letters "WAVE"
	waveTag [4]uint8 // Format 4 bytes

	// Offset 12
	// Contains the letters "fmt "
	fmtTag [4]uint8 // Subchunk1ID 4 bytes

	// Offset 16
	// 16 for PCM
	fmtLength uint32 // Subchunk1Size 4 bytes

	// Offset 20
	// PCM = 1 (i.e. Linear quantization)
	// Values other than 1 indicate some form of compression.
	audioFormat uint16 // AudioFormat 2 bytes

	// Offset 22
	// Mono = 1, Stereo = 2
	numChannels uint16 // NumChannels 2 bytes

	// Offset 24
	// 44100, 96000 etc
	sampleRate uint32 // SampleRate 4 bytes

	// Offset 28
	// = SampleRate * NumChannels * BitsPerSample/8
	byteRate uint32 // ByteRate 4 bytes

	// Offset 32
	// The number of bytes for one sample including all channels.
	// = NumChannels * BitsPerSample/8
	blockAlign uint16 // BlockAlign 2 bytes

	// Offset 34
	// 8 bits = 8, 16 bits = 16
	bitsPerSample uint16 // BitsPerSample 2 bytes

	// Offset 36
	// Contains the letters "data"
	dataTag [4]uint8 // Subchunk2ID 4 bytes

	// Offset 40
	// This is the number of bytes in the data.
	// = NumSamples * NumChannels * BitsPerSample/8
	dataLength uint32 // Subchunk2Size 4 bytes
}

func newWavfileHeader(samplesPerSecond uint32, bitsPerSample uint8, channels uint16) *wavfileHeader {
	header := &wavfileHeader{}
	copy(header.riffTag[:], "RIFF")
	copy(header.waveTag[:], "WAVE")
	copy(header.fmtTag[:], "fmt ")
	copy(header.dataTag[:], "data")
	header.riffLength = 0
	header.fmtLength = 16
	header.audioFormat = 1
	header.dataLength = 0
	header.numChannels = channels
	header.sampleRate = samplesPerSecond
	header.byteRate = samplesPerSecond * uint32(header.numChannels) * uint32(bitsPerSample/8)
	header.blockAlign = header.numChannels * uint16(bitsPerSample/8)
	header.bitsPerSample = uint16(bitsPerSample)
	return header
}

// Convert the wavfileHeader struct to a byte slice
func (h *wavfileHeader) Bytes() []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.LittleEndian, h)
	return buffer.Bytes()
}

// Create a file and return it for further writing of audio data.
func New(w io.WriterAt, samplesPerSecond uint32, bitsPerSample uint8, channels uint16, waveform []byte) error {
	header := newWavfileHeader(samplesPerSecond, bitsPerSample, channels)

	var size uint32

	written, err := w.WriteAt(header.Bytes(), 0)
	size += uint32(written)
	if err != nil {
		return err
	}

	// Write the data starting at offset 44, which is the first offset after the header.
	written, err = w.WriteAt(waveform, 44)
	size += uint32(written)
	if err != nil {
		return err
	}

	var (
		wavfileHeaderSize uint32 = 44 // bytes
		riffLength        uint32 = size - 8
		dataLength        uint32 = size - wavfileHeaderSize
	)

	// Write the riffLength into the header
	rl := make([]byte, 4)
	binary.LittleEndian.PutUint32(rl, riffLength)
	_, err = w.WriteAt(rl, 4)
	if err != nil {
		return err
	}

	// Write the length of the file into the header
	// The dataLength header starts at offset 40
	dl := make([]byte, 4)
	binary.LittleEndian.PutUint32(dl, dataLength)
	_, err = w.WriteAt(dl, 40)
	return err
}
