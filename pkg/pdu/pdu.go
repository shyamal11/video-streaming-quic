package pdu

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

const (
	// PDU types
	TYPE_DATA    = 0
	TYPE_ACK     = 1
	TYPE_VIDEO   = 2
	MAX_PDU_SIZE = 1024
)

type PDU struct {
	Mtype uint8  `json:"mtype"`
	Len   uint32 `json:"len"`
	Data  []byte `json:"data"`
	PacketNo uint32 `json:"packetNo"`
}



func MakePduBuffer() []byte {
	return make([]byte, MAX_PDU_SIZE)
}

func NewPDU(mtype uint8, packetNo uint32, data []byte) *PDU {
	return &PDU{
		Mtype: mtype,
		Len:   uint32(len(data)),
		Data:  data,
		PacketNo: packetNo,
	}
	
}

func (pdu *PDU) GetTypeAsString() string {
	switch pdu.Mtype {
	case TYPE_DATA:
		return "***DATA"
	case TYPE_ACK:
		return "****ACK"
	default:
		return "UNKNOWN"
	}
}

func (pdu *PDU) ToJsonString() string {
	jsonData, err := json.MarshalIndent(pdu, "", "    ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return "{}"
	}

	return string(jsonData)
}

func PduFromBytes(raw []byte) (*PDU, error) {
	pdu := &PDU{}
	err := json.Unmarshal(raw, pdu)
	return pdu, err
}

func (pdu *PDU) ToBytes() ([]byte, error) {
	return json.Marshal(pdu)
}

// Convert PDU to a byte slice, including a length prefix for framing
func (pdu *PDU) ToFramedBytes() ([]byte, error) {
	data, err := json.Marshal(pdu)
	if err != nil {
		return nil, err
	}

	frame := make([]byte, 4+len(data))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(data)))
	copy(frame[4:], data)
	return frame, nil
}

// Read a PDU from a stream with length prefix framing
func PduFromFramedBytes(reader io.Reader) (*PDU, error) {
	var length uint32
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	data := make([]byte, length)
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}

	return PduFromBytes(data)
}

// package pdu

// import (
// 	"encoding/json"
// 	"fmt"
// )

// const (
// 	// PDU types
// 	TYPE_DATA    = 0
// 	TYPE_ACK     = 1
// 	TYPE_VIDEO   = 2
// 	MAX_PDU_SIZE = 4096
// )

// type PDU struct {
// 	Mtype uint8  `json:"mtype"`
// 	Len   uint32 `json:"len"`
// 	Data  []byte `json:"data"`
// }

// func MakePduBuffer() []byte {
// 	return make([]byte, MAX_PDU_SIZE)
// }

// func NewPDU(mtype uint8, data []byte) *PDU {
// 	return &PDU{
// 		Mtype: mtype,
// 		Len:   uint32(len(data)),
// 		Data:  data,
// 	}
// }

// func (pdu *PDU) GetTypeAsString() string {
// 	switch pdu.Mtype {
// 	case TYPE_DATA:
// 		return "***DATA"
// 	case TYPE_ACK:
// 		return "****ACK"
// 	default:
// 		return "UNKNOWN"
// 	}
// }

// func (pdu *PDU) ToJsonString() string {
// 	jsonData, err := json.MarshalIndent(pdu, "", "    ")
// 	if err != nil {
// 		fmt.Println("Error marshaling JSON:", err)
// 		return "{}"
// 	}

// 	return string(jsonData)
// }

// func PduFromBytes(raw []byte) (*PDU, error) {
// 	pdu := &PDU{}
// 	json.Unmarshal(raw, pdu)
// 	return pdu, nil
// }

// func PduToBytes(pdu *PDU) ([]byte, error) {
// 	return json.Marshal(pdu)
// }
