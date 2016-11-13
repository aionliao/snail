package common

// Header can be converted to AudioHeaderInfo or VideoHeaderInfo
type Packet struct {
	IsVideo   bool
	Data      []byte
	TimeStamp uint32 // dts
	Header    PacketHeader
}
