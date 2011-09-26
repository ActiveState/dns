// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Extended and bugfixes by Miek Gieben

package dns

import (
	"os"
	"net"
	"time"
	"strconv"
	"strings"
)

// Packet formats

// Wire constants and supported types.
const (
	// valid RR_Header.Rrtype and Question.qtype
	TypeA     uint16 = 1
	TypeNS    uint16 = 2
	TypeMD    uint16 = 3
	TypeMF    uint16 = 4
	TypeCNAME uint16 = 5
	TypeSOA   uint16 = 6
	TypeMB    uint16 = 7
	TypeMG    uint16 = 8
	TypeMR    uint16 = 9
	TypeNULL  uint16 = 10
	TypeWKS   uint16 = 11
	TypePTR   uint16 = 12
	TypeHINFO uint16 = 13
	TypeMINFO uint16 = 14
	TypeMX    uint16 = 15
	TypeTXT   uint16 = 16
	TypeAAAA  uint16 = 28
	TypeLOC   uint16 = 29
	TypeSRV   uint16 = 33
	TypeNAPTR uint16 = 35
	TypeKX    uint16 = 36
	TypeCERT  uint16 = 37
	TypeDNAME uint16 = 39

	// EDNS
	TypeOPT uint16 = 41

	TypeSIG        uint16 = 24
	TypeKEY        uint16 = 25
	TypeNXT        uint16 = 30
	TypeDS         uint16 = 43
	TypeSSHFP      uint16 = 44
	TypeIPSECKEY   uint16 = 45 // No type implemented
	TypeRRSIG      uint16 = 46
	TypeNSEC       uint16 = 47
	TypeDNSKEY     uint16 = 48
	TypeDHCID      uint16 = 49
	TypeNSEC3      uint16 = 50
	TypeNSEC3PARAM uint16 = 51
	TypeTALINK     uint16 = 58
	TypeSPF        uint16 = 99

	TypeTKEY uint16 = 249
	TypeTSIG uint16 = 250
	// valid Question.Qtype only
	TypeIXFR  uint16 = 251
	TypeAXFR  uint16 = 252
	TypeMAILB uint16 = 253
	TypeMAILA uint16 = 254
	TypeANY   uint16 = 255
	TypeURI   uint16 = 256
	TypeTA    uint16 = 32768
	TypeDLV   uint16 = 32769

	TypeNSEC4      uint16 = 65326
	TypeNSEC4PARAM uint16 = 65327

	// valid Question.Qclass
	ClassINET   = 1
	ClassCSNET  = 2
	ClassCHAOS  = 3
	ClassHESIOD = 4
	ClassNONE   = 254
	ClassANY    = 255

	// Msg.rcode
	RcodeSuccess        = 0
	RcodeFormatError    = 1
	RcodeServerFailure  = 2
	RcodeNameError      = 3
	RcodeNotImplemented = 4
	RcodeRefused        = 5
	RcodeYXDomain       = 6
	RcodeYXRrset        = 7
	RcodeNXRrset        = 8
	RcodeNotAuth        = 9
	RcodeNotZone        = 10
	RcodeBadSig         = 16 // TSIG
	RcodeBadKey         = 17
	RcodeBadTime        = 18
	RcodeBadMode        = 19 // TKEY
	RcodeBadName        = 20
	RcodeBadAlg         = 21
	RcodeBadTrunc       = 22 // TSIG

	// Opcode
	OpcodeQuery  = 0
	OpcodeIQuery = 1
	OpcodeStatus = 2
	// There is no 3
	OpcodeNotify = 4
	OpcodeUpdate = 5
)

// The wire format for the DNS packet header.
type Header struct {
	Id                                 uint16
	Bits                               uint16
	Qdcount, Ancount, Nscount, Arcount uint16
}

const (
	// Header.Bits
	_QR = 1 << 15 // query/response (response=1)
	_AA = 1 << 10 // authoritative
	_TC = 1 << 9  // truncated
	_RD = 1 << 8  // recursion desired
	_RA = 1 << 7  // recursion available
	_Z  = 1 << 6  // Z
	_AD = 1 << 5  // authticated data
	_CD = 1 << 4  // checking disabled
)

// DNS queries.
type Question struct {
	Name   string "domain-name" // "domain-name" specifies encoding
	Qtype  uint16
	Qclass uint16
}

func (q *Question) String() string {
	// prefix with ; (as in dig)
	s := ";" + q.Name + "\t"
	s = s + Class_str[q.Qclass] + "\t"
	if _, ok := Rr_str[q.Qtype]; ok {
		s += " " + Rr_str[q.Qtype]
	} else {
		s += " " + "TYPE" + strconv.Itoa(int(q.Qtype))
	}
	return s
}

// NewRRString returns the last RR contained in s.
func NewRRString(s string) (RR, os.Error) {
	p := NewParser(strings.NewReader(s))
	return p.First()
}

// NewRR returns a new RR with the hdr.Rrtype also set.
// If the type i is not known, nil is returned.
func NewRR(i uint16) RR {
	r := rr_mk[i]()
	r.Header().Rrtype = i
	return r
}

type RR_ANY struct {
	Hdr RR_Header
	// Does not have any rdata
}

func (rr *RR_ANY) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_ANY) String() string {
	return rr.Hdr.String()
}

type RR_CNAME struct {
	Hdr   RR_Header
	Cname string "domain-name"
}

func (rr *RR_CNAME) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_CNAME) String() string {
	return rr.Hdr.String() + rr.Cname
}

type RR_HINFO struct {
	Hdr RR_Header
	Cpu string
	Os  string
}

func (rr *RR_HINFO) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_HINFO) String() string {
	return rr.Hdr.String() + rr.Cpu + " " + rr.Os
}

type RR_MB struct {
	Hdr RR_Header
	Mb  string "domain-name"
}

func (rr *RR_MB) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_MB) String() string {
	return rr.Hdr.String() + rr.Mb
}

type RR_MG struct {
	Hdr RR_Header
	Mg  string "domain-name"
}

func (rr *RR_MG) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_MG) String() string {
	return rr.Hdr.String() + rr.Mg
}

type RR_MINFO struct {
	Hdr   RR_Header
	Rmail string "domain-name"
	Email string "domain-name"
}

func (rr *RR_MINFO) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_MINFO) String() string {
	return rr.Hdr.String() + rr.Rmail + " " + rr.Email
}

type RR_MR struct {
	Hdr RR_Header
	Mr  string "domain-name"
}

func (rr *RR_MR) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_MR) String() string {
	return rr.Hdr.String() + rr.Mr
}

type RR_MX struct {
	Hdr  RR_Header
	Pref uint16
	Mx   string "domain-name"
}

func (rr *RR_MX) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_MX) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Pref)) + " " + rr.Mx
}

type RR_NS struct {
	Hdr RR_Header
	Ns  string "domain-name"
}

func (rr *RR_NS) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_NS) String() string {
	return rr.Hdr.String() + rr.Ns
}

type RR_PTR struct {
	Hdr RR_Header
	Ptr string "domain-name"
}

func (rr *RR_PTR) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_PTR) String() string {
	return rr.Hdr.String() + rr.Ptr
}

type RR_SOA struct {
	Hdr     RR_Header
	Ns      string "domain-name"
	Mbox    string "domain-name"
	Serial  uint32
	Refresh uint32
	Retry   uint32
	Expire  uint32
	Minttl  uint32
}

func (rr *RR_SOA) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_SOA) String() string {
	return rr.Hdr.String() + rr.Ns + " " + rr.Mbox +
		" " + strconv.Itoa(int(rr.Serial)) +
		" " + strconv.Itoa(int(rr.Refresh)) +
		" " + strconv.Itoa(int(rr.Retry)) +
		" " + strconv.Itoa(int(rr.Expire)) +
		" " + strconv.Itoa(int(rr.Minttl))
}

type RR_TXT struct {
	Hdr RR_Header
	Txt string "txt"
}

func (rr *RR_TXT) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_TXT) String() string {
	return rr.Hdr.String() + "\"" + rr.Txt + "\""
}

type RR_SRV struct {
	Hdr      RR_Header
	Priority uint16
	Weight   uint16
	Port     uint16
	Target   string "domain-name"
}

func (rr *RR_SRV) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_SRV) String() string {
	return rr.Hdr.String() +
		strconv.Itoa(int(rr.Priority)) + " " +
		strconv.Itoa(int(rr.Weight)) + " " +
		strconv.Itoa(int(rr.Port)) + " " + rr.Target
}

type RR_NAPTR struct {
	Hdr         RR_Header
	Order       uint16
	Preference  uint16
	Flags       string
	Service     string
	Regexp      string
	Replacement string "domain-name"
}

func (rr *RR_NAPTR) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_NAPTR) String() string {
	return rr.Hdr.String() +
		strconv.Itoa(int(rr.Order)) + " " +
		strconv.Itoa(int(rr.Preference)) + " " +
		"\"" + rr.Flags + "\" " +
		"\"" + rr.Service + "\" " +
		"\"" + rr.Regexp + "\" " +
		rr.Replacement
}

// See RFC 4398.
type RR_CERT struct {
	Hdr         RR_Header
	Type        uint16
	KeyTag      uint16
	Algorithm   uint8
	Certificate string "base64"
}

func (rr *RR_CERT) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_CERT) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Type)) +
		" " + strconv.Itoa(int(rr.KeyTag)) +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + rr.Certificate
}

// See RFC 2672.
type RR_DNAME struct {
	Hdr    RR_Header
	Target string "domain-name"
}

func (rr *RR_DNAME) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_DNAME) String() string {
	return rr.Hdr.String() + rr.Target
}

type RR_A struct {
	Hdr RR_Header
	A   net.IP "A"
}

func (rr *RR_A) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_A) String() string {
	return rr.Hdr.String() + rr.A.String()
}

type RR_AAAA struct {
	Hdr  RR_Header
	AAAA net.IP "AAAA"
}

func (rr *RR_AAAA) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_AAAA) String() string {
	return rr.Hdr.String() + rr.AAAA.String()
}

type RR_LOC struct {
	Hdr       RR_Header
	Version   uint8
	Size      uint8
	HorizPre  uint8
	VertPre   uint8
	Latitude  uint32
	Longitude uint32
	Altitude  uint32
}

func (rr *RR_LOC) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_LOC) String() string {
	// Version is not shown
	return rr.Hdr.String() + "TODO"
}

type RR_RRSIG struct {
	Hdr         RR_Header
	TypeCovered uint16
	Algorithm   uint8
	Labels      uint8
	OrigTtl     uint32
	Expiration  uint32
	Inception   uint32
	KeyTag      uint16
	SignerName  string "domain-name"
	Signature   string "base64"
}

func (rr *RR_RRSIG) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_RRSIG) String() string {
	return rr.Hdr.String() + Rr_str[rr.TypeCovered] +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + strconv.Itoa(int(rr.Labels)) +
		" " + strconv.Itoa(int(rr.OrigTtl)) +
		" " + timeToDate(rr.Expiration) +
		" " + timeToDate(rr.Inception) +
		" " + strconv.Itoa(int(rr.KeyTag)) +
		" " + rr.SignerName +
		" " + rr.Signature
}

type RR_NSEC struct {
	Hdr        RR_Header
	NextDomain string   "domain-name"
	TypeBitMap []uint16 "NSEC"
}

func (rr *RR_NSEC) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_NSEC) String() string {
	s := rr.Hdr.String() + rr.NextDomain
	for i := 0; i < len(rr.TypeBitMap); i++ {
		if _, ok := Rr_str[rr.TypeBitMap[i]]; ok {
			s += " " + Rr_str[rr.TypeBitMap[i]]
		} else {
			s += " " + "TYPE" + strconv.Itoa(int(rr.TypeBitMap[i]))
		}
	}
	return s
}

type RR_DS struct {
	Hdr        RR_Header
	KeyTag     uint16
	Algorithm  uint8
	DigestType uint8
	Digest     string "hex"
}

func (rr *RR_DS) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_DS) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.KeyTag)) +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + strconv.Itoa(int(rr.DigestType)) +
		" " + strings.ToUpper(rr.Digest)
}

type RR_DLV struct {
	Hdr        RR_Header
	KeyTag     uint16
	Algorithm  uint8
	DigestType uint8
	Digest     string "hex"
}

func (rr *RR_DLV) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_DLV) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.KeyTag)) +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + strconv.Itoa(int(rr.DigestType)) +
		" " + strings.ToUpper(rr.Digest)
}

type RR_KX struct {
	Hdr        RR_Header
	Preference uint16
	Exchanger  string "domain-name"
}

func (rr *RR_KX) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_KX) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Preference)) +
		" " + rr.Exchanger
}

type RR_TA struct {
	Hdr        RR_Header
	KeyTag     uint16
	Algorithm  uint8
	DigestType uint8
	Digest     string "hex"
}

func (rr *RR_TA) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_TA) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.KeyTag)) +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + strconv.Itoa(int(rr.DigestType)) +
		" " + strings.ToUpper(rr.Digest)
}

type RR_TALINK struct {
	Hdr          RR_Header
	PreviousName string "domain"
	NextName     string "domain"
}

func (rr *RR_TALINK) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_TALINK) String() string {
	return rr.Hdr.String() +
		" " + rr.PreviousName + " " + rr.NextName
}

type RR_SSHFP struct {
	Hdr         RR_Header
	Algorithm   uint8
	Type        uint8
	FingerPrint string "hex"
}

func (rr *RR_SSHFP) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_SSHFP) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Algorithm)) +
		" " + strconv.Itoa(int(rr.Type)) +
		" " + strings.ToUpper(rr.FingerPrint)
}

type RR_DNSKEY struct {
	Hdr       RR_Header
	Flags     uint16
	Protocol  uint8
	Algorithm uint8
	PublicKey string "base64"
}

func (rr *RR_DNSKEY) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_DNSKEY) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Flags)) +
		" " + strconv.Itoa(int(rr.Protocol)) +
		" " + strconv.Itoa(int(rr.Algorithm)) +
		" " + rr.PublicKey
}

type RR_NSEC3 struct {
	Hdr        RR_Header
	Hash       uint8
	Flags      uint8
	Iterations uint16
	SaltLength uint8
	Salt       string "size-hex"
	HashLength uint8
	NextDomain string   "size-base32"
	TypeBitMap []uint16 "NSEC"
}

func (rr *RR_NSEC3) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_NSEC3) String() string {
	s := rr.Hdr.String()
	s += strconv.Itoa(int(rr.Hash)) +
		" " + strconv.Itoa(int(rr.Flags)) +
		" " + strconv.Itoa(int(rr.Iterations)) +
		" " + strings.ToUpper(rr.Salt) +
		" " + rr.NextDomain
	for i := 0; i < len(rr.TypeBitMap); i++ {
		if _, ok := Rr_str[rr.TypeBitMap[i]]; ok {
			s += " " + Rr_str[rr.TypeBitMap[i]]
		} else {
			s += " " + "TYPE" + strconv.Itoa(int(rr.TypeBitMap[i]))
		}
	}
	return s
}

type RR_NSEC3PARAM struct {
	Hdr        RR_Header
	Hash       uint8
	Flags      uint8
	Iterations uint16
	SaltLength uint8
	Salt       string "hex" // hexsize??
}

func (rr *RR_NSEC3PARAM) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_NSEC3PARAM) String() string {
	s := rr.Hdr.String()
	s += strconv.Itoa(int(rr.Hash)) +
		" " + strconv.Itoa(int(rr.Flags)) +
		" " + strconv.Itoa(int(rr.Iterations)) +
		" " + strings.ToUpper(rr.Salt)
	return s
}

// See RFC 4408.
type RR_SPF struct {
	Hdr RR_Header
	Txt string
}

func (rr *RR_SPF) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_SPF) String() string {
	return rr.Hdr.String() + "\"" + rr.Txt + "\""
}

type RR_TKEY struct {
	Hdr        RR_Header
	Algorithm  string "domain-name"
	Inception  uint32
	Expiration uint32
	Mode       uint16
	Error      uint16
	KeySize    uint16
	Key        string
	Otherlen   uint16
	OtherData  string
}

func (rr *RR_TKEY) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_TKEY) String() string {
	// It has no presentation format
	return ""
}

// Unknown RR representation
type RR_RFC3597 struct {
	Hdr   RR_Header
	Rdata string "hex"
}

func (rr *RR_RFC3597) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_RFC3597) String() string {
	s := rr.Hdr.String()
	s += "\\# " + strconv.Itoa(len(rr.Rdata)/2) + " " + rr.Rdata
	return s
}

type RR_URI struct {
	Hdr      RR_Header
	Priority uint16
	Weight   uint16
	Target   string
}

func (rr *RR_URI) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_URI) String() string {
	return rr.Hdr.String() + strconv.Itoa(int(rr.Priority)) +
		" " + strconv.Itoa(int(rr.Weight)) +
		" " + rr.Target
}

type RR_DHCID struct {
	Hdr    RR_Header
	Digest string "base64"
}

func (rr *RR_DHCID) Header() *RR_Header {
	return &rr.Hdr
}

func (rr *RR_DHCID) String() string {
	return rr.Hdr.String() + rr.Digest
}

// RFC 2845.
type RR_TSIG struct {
	Hdr        RR_Header
	Algorithm  string "domain-name"
	TimeSigned uint64
	Fudge      uint16
	MACSize    uint16
	MAC        string "size-hex"
	OrigId     uint16
	Error      uint16
	OtherLen   uint16
	OtherData  string "size-hex"
}

func (rr *RR_TSIG) Header() *RR_Header {
	return &rr.Hdr
}

// TSIG has no official presentation format, but this will suffice.
func (rr *RR_TSIG) String() string {
	return rr.Hdr.String() +
		" " + rr.Algorithm +
		" " + tsigTimeToDate(rr.TimeSigned) +
		" " + strconv.Itoa(int(rr.Fudge)) +
		" " + strconv.Itoa(int(rr.MACSize)) +
		" " + strings.ToUpper(rr.MAC) +
		" " + strconv.Itoa(int(rr.OrigId)) +
		" " + strconv.Itoa(int(rr.Error)) + // BIND prints NOERROR
		" " + strconv.Itoa(int(rr.OtherLen)) +
		" " + rr.OtherData
}

// Translate the RRSIG's incep. and expir. time to the correct date.
// Taking into account serial arithmetic (RFC 1982)
func timeToDate(t uint32) string {
	utc := time.UTC().Seconds()
	mod := (int64(t) - utc) / Year68

	// If needed assume wrap around(s)
	ti := time.SecondsToUTC(int64(t) + (mod * Year68)) // abs()? TODO
	return ti.Format("20060102150405")
}

// Translate the TSIG time signed into a date. There is no
// need for RFC1982 calculations as this date is 48 bits
func tsigTimeToDate(t uint64) string {
	// only use the lower 48 bits, TODO(mg), check for 48 bit size
	ti := time.SecondsToUTC(int64(t))
	return ti.Format("20060102150405")
}

// Map of constructors for each RR wire type.
var rr_mk = map[uint16]func() RR{
	TypeCNAME:      func() RR { return new(RR_CNAME) },
	TypeHINFO:      func() RR { return new(RR_HINFO) },
	TypeMB:         func() RR { return new(RR_MB) },
	TypeMG:         func() RR { return new(RR_MG) },
	TypeMINFO:      func() RR { return new(RR_MINFO) },
	TypeMR:         func() RR { return new(RR_MR) },
	TypeMX:         func() RR { return new(RR_MX) },
	TypeNS:         func() RR { return new(RR_NS) },
	TypePTR:        func() RR { return new(RR_PTR) },
	TypeSOA:        func() RR { return new(RR_SOA) },
	TypeTXT:        func() RR { return new(RR_TXT) },
	TypeSRV:        func() RR { return new(RR_SRV) },
	TypeNAPTR:      func() RR { return new(RR_NAPTR) },
	TypeDNAME:      func() RR { return new(RR_DNAME) },
	TypeA:          func() RR { return new(RR_A) },
	TypeAAAA:       func() RR { return new(RR_AAAA) },
	TypeLOC:        func() RR { return new(RR_LOC) },
	TypeOPT:        func() RR { return new(RR_OPT) },
	TypeDS:         func() RR { return new(RR_DS) },
	TypeCERT:       func() RR { return new(RR_CERT) },
	TypeKX:         func() RR { return new(RR_KX) },
	TypeSPF:        func() RR { return new(RR_SPF) },
	TypeTALINK:     func() RR { return new(RR_TALINK) },
	TypeSSHFP:      func() RR { return new(RR_SSHFP) },
	TypeRRSIG:      func() RR { return new(RR_RRSIG) },
	TypeNSEC:       func() RR { return new(RR_NSEC) },
	TypeDNSKEY:     func() RR { return new(RR_DNSKEY) },
	TypeNSEC3:      func() RR { return new(RR_NSEC3) },
	TypeDHCID:      func() RR { return new(RR_DHCID) },
	TypeNSEC3PARAM: func() RR { return new(RR_NSEC3PARAM) },
	TypeTKEY:       func() RR { return new(RR_TKEY) },
	TypeTSIG:       func() RR { return new(RR_TSIG) },
	TypeURI:        func() RR { return new(RR_URI) },
	TypeTA:         func() RR { return new(RR_TA) },
	TypeDLV:        func() RR { return new(RR_DLV) },
}
