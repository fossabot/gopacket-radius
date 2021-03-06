package radius

import (
	"reflect"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func checkLayers(p gopacket.Packet, want []gopacket.LayerType, t *testing.T) {
	layers := p.Layers()
	t.Log("Checking packet layers, want", want)
	for _, l := range layers {
		t.Logf("  Got layer %v, %d bytes, payload of %d bytes", l.LayerType(),
			len(l.LayerContents()), len(l.LayerPayload()))
	}
	t.Log(p)
	if len(layers) < len(want) {
		t.Errorf("  Number of layers mismatch: got %d want %d", len(layers),
			len(want))
		return
	}
	for i, l := range want {
		if l == gopacket.LayerTypePayload {
			// done matching layers
			return
		}

		if layers[i].LayerType() != l {
			t.Errorf("  Layer %d mismatch: got %v want %v", i,
				layers[i].LayerType(), l)
		}
	}
}

func checkRADIUS(desc string, t *testing.T, packetBytes []byte, pExpectedRADIUS *RADIUS) {
	// Analyse the packet bytes, yielding a new packet object p.
	p := gopacket.NewPacket(packetBytes, layers.LinkTypeEthernet, gopacket.Default)
	if p.ErrorLayer() != nil {
		t.Errorf("Failed to decode packet %s: %v", desc, p.ErrorLayer().Error())
	}

	// Ensure that the packet analysis yielded the correct set of layers:
	//    Link Layer        = Ethernet.
	//    Network Layer     = IPv4.
	//    Transport Layer   = UDP.
	//    Application Layer = RADIUS.
	checkLayers(p, []gopacket.LayerType{
		layers.LayerTypeEthernet,
		layers.LayerTypeIPv4,
		layers.LayerTypeUDP,
		LayerTypeRADIUS,
	}, t)

	// Select the Application (RADIUS) layer.
	pResultRADIUS, ok := p.ApplicationLayer().(*RADIUS)
	if !ok {
		t.Error("No RADIUS layer type found in packet in " + desc + ".")
	}

	// Compare the generated RADIUS object with the expected RADIUS object.
	if !reflect.DeepEqual(pResultRADIUS, pExpectedRADIUS) {
		t.Errorf("RADIUS packet processing failed for packet "+desc+
			":\ngot  :\n%#v\n\nwant :\n%#v\n\n", pResultRADIUS, pExpectedRADIUS)
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err := pResultRADIUS.SerializeTo(buf, opts)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(pResultRADIUS.BaseLayer.Contents, buf.Bytes()) {
		t.Errorf("RADIUS packet serialization failed for packet "+desc+
			":\ngot  :\n%x\n\nwant :\n%x\n\n", buf.Bytes(), packetBytes)
	}
}

func TestRADIUSAccessRequest(t *testing.T) {
	// This test packet is the first RADIUS packet in the RADIUS sample capture
	// pcap file radtest.pcap on the Wireshark sample captures page:
	//
	//    https://github.com/egxp/docker-compose-test-radius
	var testPacketRADIUS = []byte{
		0x02, 0x42, 0xac, 0x14, 0x00, 0x02, 0x02, 0x42, 0x06, 0x4d, 0xad, 0xbf, 0x08, 0x00, 0x45, 0x00,
		0x00, 0x67, 0xee, 0xea, 0x40, 0x00, 0x40, 0x11, 0xf3, 0x6f, 0xac, 0x14, 0x00, 0x01, 0xac, 0x14,
		0x00, 0x02, 0xd8, 0x29, 0x07, 0x14, 0x00, 0x53, 0x58, 0x90, 0x01, 0x8d, 0x00, 0x4b, 0x3b, 0xbd,
		0x22, 0x52, 0xb4, 0xc8, 0xd8, 0x44, 0x1b, 0x46, 0x79, 0xbf, 0x4a, 0x2b, 0x86, 0x01, 0x01, 0x07,
		0x41, 0x64, 0x6d, 0x69, 0x6e, 0x02, 0x12, 0x4d, 0x2f, 0x62, 0x0b, 0x33, 0x9d, 0x6d, 0x1f, 0xe0,
		0xe4, 0x6d, 0x1f, 0x9b, 0xda, 0xff, 0xf0, 0x04, 0x06, 0x7f, 0x00, 0x01, 0x01, 0x05, 0x06, 0x00,
		0x00, 0x00, 0x00, 0x50, 0x12, 0x41, 0x73, 0xed, 0x26, 0xd3, 0xb3, 0xa9, 0x64, 0xff, 0x4d, 0xc3,
		0x0d, 0x94, 0x33, 0xe8, 0x2a,
	}

	// Assemble the RADIUS object that we expect to emerge from this test.
	pExpectedRADIUS := &RADIUS{
		BaseLayer: layers.BaseLayer{
			Contents: []byte{
				0x01, 0x8d, 0x00, 0x4b, 0x3b, 0xbd, 0x22, 0x52, 0xb4, 0xc8, 0xd8, 0x44, 0x1b, 0x46, 0x79, 0xbf,
				0x4a, 0x2b, 0x86, 0x01, 0x01, 0x07, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x02, 0x12, 0x4d, 0x2f, 0x62,
				0x0b, 0x33, 0x9d, 0x6d, 0x1f, 0xe0, 0xe4, 0x6d, 0x1f, 0x9b, 0xda, 0xff, 0xf0, 0x04, 0x06, 0x7f,
				0x00, 0x01, 0x01, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x50, 0x12, 0x41, 0x73, 0xed, 0x26, 0xd3,
				0xb3, 0xa9, 0x64, 0xff, 0x4d, 0xc3, 0x0d, 0x94, 0x33, 0xe8, 0x2a,
			},
			Payload: nil,
		},
		Code:       RADIUSCodeAccessRequest,
		Identifier: RADIUSIdentifier(0x8d),
		Length:     RADIUSLength(0x004b),
		Authenticator: RADIUSAuthenticator([16]byte{
			0x3b, 0xbd, 0x22, 0x52, 0xb4, 0xc8, 0xd8, 0x44, 0x1b, 0x46, 0x79, 0xbf, 0x4a, 0x2b, 0x86, 0x01,
		}),
		Attributes: []RADIUSAttribute{
			{
				Type:   RADIUSAttributeTypeUserName,
				Length: RADIUSAttributeLength(0x07),
				Value:  RADIUSAttributeValue("Admin"),
			},
			{
				Type:   RADIUSAttributeTypeUserPassword,
				Length: RADIUSAttributeLength(0x12),
				Value:  RADIUSAttributeValue("\x4d\x2f\x62\x0b\x33\x9d\x6d\x1f\xe0\xe4\x6d\x1f\x9b\xda\xff\xf0"),
			},
			{
				Type:   RADIUSAttributeTypeNASIPAddress,
				Length: RADIUSAttributeLength(0x06),
				Value:  RADIUSAttributeValue("\x7f\x00\x01\x01"),
			},
			{
				Type:   RADIUSAttributeTypeNASPort,
				Length: RADIUSAttributeLength(0x06),
				Value:  RADIUSAttributeValue("\x00\x00\x00\x00"),
			},
			{
				Type:   RADIUSAttributeTypeMessageAuthenticator,
				Length: RADIUSAttributeLength(0x12),
				Value:  RADIUSAttributeValue("\x41\x73\xed\x26\xd3\xb3\xa9\x64\xff\x4d\xc3\x0d\x94\x33\xe8\x2a"),
			},
		},
	}

	checkRADIUS("AccessRequest", t, testPacketRADIUS, pExpectedRADIUS)
}

func TestRADIUSAccessAccept(t *testing.T) {
	// This test packet is the first RADIUS packet in the RADIUS sample capture
	// pcap file radtest.pcap on the Wireshark sample captures page:
	//
	//    https://github.com/egxp/docker-compose-test-radius
	var testPacketRADIUS = []byte{
		0x02, 0x42, 0x06, 0x4d, 0xad, 0xbf, 0x02, 0x42, 0xac, 0x14, 0x00, 0x02, 0x08, 0x00, 0x45, 0x00,
		0x00, 0x30, 0xee, 0xfd, 0x00, 0x00, 0x40, 0x11, 0x33, 0x94, 0xac, 0x14, 0x00, 0x02, 0xac, 0x14,
		0x00, 0x01, 0x07, 0x14, 0xd8, 0x29, 0x00, 0x1c, 0x58, 0x59, 0x02, 0x8d, 0x00, 0x14, 0x86, 0xa8,
		0xd5, 0xcd, 0x69, 0x3c, 0x07, 0x5e, 0x9e, 0x18, 0xa2, 0x2d, 0xdd, 0x5f, 0x2b, 0xff,
	}

	// Assemble the RADIUS object that we expect to emerge from this test.
	pExpectedRADIUS := &RADIUS{
		BaseLayer: layers.BaseLayer{
			Contents: []byte{
				0x02, 0x8d, 0x00, 0x14, 0x86, 0xa8, 0xd5, 0xcd, 0x69, 0x3c, 0x07, 0x5e, 0x9e, 0x18, 0xa2, 0x2d,
				0xdd, 0x5f, 0x2b, 0xff,
			},
			Payload: nil,
		},
		Code:       RADIUSCodeAccessAccept,
		Identifier: RADIUSIdentifier(0x8d),
		Length:     RADIUSLength(0x0014),
		Authenticator: RADIUSAuthenticator([16]byte{
			0x86, 0xa8, 0xd5, 0xcd, 0x69, 0x3c, 0x07, 0x5e, 0x9e, 0x18, 0xa2, 0x2d, 0xdd, 0x5f, 0x2b, 0xff,
		}),
	}

	checkRADIUS("AccessAccept", t, testPacketRADIUS, pExpectedRADIUS)
}
