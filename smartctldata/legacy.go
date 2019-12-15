package smartctldata

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type lineParser interface {
	Parse(*Output, string) (lineParser, error)
}

type lineParserFunc func(*Output, string) (lineParser, error)

func (f lineParserFunc) Parse(o *Output, l string) (lineParser, error) { return f(o, l) }

var topParsers = map[string]lineParser{
	"=== START OF INFORMATION SECTION ===":     lineParserFunc(parseInfo),
	"=== START OF READ SMART DATA SECTION ===": lineParserFunc(parseSMARTData),
	"=== SMARTCTL2PROM ===":                    lineParserFunc(parseSmartCtl2Prom),
	"=== END ===":                              nil,
}

func parseSMARTCtl(r *bufio.Reader, o *Output) error {
	var parser lineParser
	for done, input := false, false; !done; {
		l, err := r.ReadString('\n')
		if err != nil {
			if err != io.EOF || !input {
				return err
			}
			done = true
		}
		input = true
		l = strings.TrimSpace(l)
		if parser == nil {
			var ok bool
			parser, ok = topParsers[l]
			if ok && parser == nil {
				done = true
			}
			continue

		}
		parser, err = parser.Parse(o, l)
		if err != nil {
			return err
		}
	}

	// smart_device_interface_speed_bps{device_name="/dev/ada0",device_serial_number="WD-WMC300098101"} 6e+09
	// o.InterfaceSpeed.Current.UnitsPerSecond * o.InterfaceSpeed.Current.BitsPerUnit

	// "local_time": { "time_t": 1561919685, "asctime": "Sun Jun 30 18:34:45 2019 UTC" },
	// smart_device_read_time{device_model_family="Western Digital Red",device_model_name="WDC WD20EFRX-68AX9N0",device_name="/dev/ada0",device_serial_number="WD-WMC300098101",device_type="atacam",smartctl_exit_status="0"} 1.561919685e+09

	for i := range o.ATASMARTAttributes.Table {
		switch rawValue := o.ATASMARTAttributes.Table[i].Raw.Value; o.ATASMARTAttributes.Table[i].ID {
		case 9:
			o.PowerOnTime.Hours = rawValue
		case 12:
			o.PowerCycleCount = rawValue
		case 194:
			o.Temperature.Current = rawValue & 0x00000000ffff
		}
	}

	return nil
}

const smartCtlDate = "Mon Jan _2 15:04:05 2006 MST"

func parseInfo(o *Output, l string) (lineParser, error) {
	if l == "" {
		return nil, nil
	}
	f := strings.SplitN(l, ":", 2)
	if len(f) != 2 {
		return nil, fmt.Errorf("parseInfo: cannot split %q into key:value pair", l)
	} else {
		switch k, v := strings.ToLower(strings.TrimSpace(f[0])), strings.TrimSpace(f[1]); k {
		case "model family":
			o.ModelFamily = v
		case "device model":
			o.ModelName = v
		case "serial number":
			o.SerialNumber = v
		case "lu wwn device id":
			f := strings.SplitN(v, " ", 3)
			if len(f) != 3 {
				return nil, fmt.Errorf("parseInfo: cannot split %q into NAA, OUI and ID", v)
			}
			for i, p := range []*int64{&o.WWN.NAA, &o.WWN.OUI, &o.WWN.ID} {
				n, err := strconv.ParseInt(f[i], 16, 64)
				if err != nil {
					return nil, fmt.Errorf("parseInfo: cannot parse %q as hexadecimal number: %v", f[i], err)
				}
				*p = n
			}
		case "firmware version":
			o.FirmwareVersion = v
		case "user capacity":
			// JSON: "user_capacity": { "blocks": 3907029168, "bytes": 2000398934016 },
			// Text: User Capacity    2,000,398,934,016 bytes [2.00 TB]
			if i := strings.Index(v, " bytes ["); i < 0 {
				return nil, fmt.Errorf("parseInfo: cannot parse %q as capacity", v)
			} else {
				v = strings.ReplaceAll(v[:i], ",", "")
			}
			if bytes, err := strconv.ParseInt(v, 10, 64); err != nil {
				return nil, fmt.Errorf("parseInfo: cannot parse %q as decimal number: %v", v, err)
			} else {
				o.UserCapacity.Bytes = bytes
			}
			if o.LogicalBlockSize != 0 {
				o.UserCapacity.Blocks = o.UserCapacity.Bytes / o.LogicalBlockSize
			}
		case "sector sizes":
			// JSON: "logical_block_size": 512, "physical_block_size": 4096
			// Text: Sector Sizes     512 bytes logical, 4096 bytes physical
			f := strings.SplitN(v, ", ", 2)
			for _, s := range f {
				ss := strings.SplitN(s, " ", 2)
				if len(ss) != 2 {
					return nil, fmt.Errorf("parseInfo: cannot parse %q as block size", s)
				}
				var p *int64
				switch ss[1] {
				case "bytes logical":
					p = &o.LogicalBlockSize
				case "bytes physical":
					p = &o.PhysicalBlockSize
				default:
					return nil, fmt.Errorf("parseInfo: cannot parse %q as block size", s)
				}
				if n, err := strconv.ParseInt(ss[0], 10, 64); err != nil {
					return nil, fmt.Errorf("parseInfo: cannot parse %q as decimal number: %v", ss[0], err)
				} else {
					*p = n
				}
			}
			if o.LogicalBlockSize != 0 {
				o.UserCapacity.Blocks = o.UserCapacity.Bytes / o.LogicalBlockSize
			}
		case "device is":
			// JSON: "in_smartctl_database": true
			// Text: Device is        In smartctl database [for details use: -P show]
			o.InSmartCtlDatabase = strings.HasPrefix(v, "In smartctl database")
		case "ata Version is":
			// JSON "ata_version": { "string": "ACS-2 (minor revision not indicated)", "major_value": 1022, "minor_value": 0 },
			// ATA Version is   ACS-2 (minor revision not indicated)
			// JSON "sata_version": { "string": "SATA 3.0", "value": 62 }, "interface_speed": {"max": {} "current": {}}
			// SATA Version is  SATA 3.0, 6.0 Gb/s (current: 6.0 Gb/s)
		case "local time is":
			o.LocalTime.AscTime = v
			// JSON "local_time": { "time_t": 1561919685, "asctime": "Sun Jun 30 18:34:45 2019 UTC" },
			// Text Local Time is    Mon Jul  1 16:57:20 2019 UTC
			if t, err := time.Parse(smartCtlDate, v); err != nil {
				return nil, fmt.Errorf("parseInfo: %v", v, err)
			} else {
				o.LocalTime.TimeT = t.Unix()
			}

			// case k == "smart support is":
			// 	switch v {
			// 	case "Available - device has SMART capability.":
			// 		o.smartAvailable = true
			// 	case "Enabled":
			// 		o.smartEnabled = true
			// 	}
		}
	}
	return lineParserFunc(parseInfo), nil
}

func parseSMARTData(o *Output, l string) (lineParser, error) {
	if l == "Vendor Specific SMART Attributes with Thresholds:" {
		return &parseSMARTAttrs{}, nil
	}
	if f := strings.SplitN(l, ":", 2); len(f) == 2 && f[0] == "SMART overall-health self-assessment test result" {
		// JSON:   "smart_status": { "passed": true },
		o.SMARTStatus.Passed = strings.EqualFold(strings.TrimSpace(f[1]), "PASSED")
	}
	return lineParserFunc(parseSMARTData), nil
}

type indices struct {
	name, flag, raw_value, threshold, value, worst int
}

func (i *indices) allSet() bool {
	return i.name != 0 && i.value != 0 && i.worst != 0 && i.threshold != 0 && i.raw_value != 0
}

type parseSMARTAttrs struct {
	idx *indices
}

func (p *parseSMARTAttrs) Parse(o *Output, l string) (lineParser, error) {
	if l == "" {
		return nil, nil
	}
	fs := strings.Fields(l)
	if fs[0] == "ID#" {
		if p.idx != nil {
			return nil, fmt.Errorf("parseSMARTAttrs: repeated header lines in SMART Attributes section")
		}
		p.idx = &indices{}
		fieldMap := map[string]*int{
			"ATTRIBUTE_NAME": &p.idx.name,
			"FLAG":           &p.idx.flag,
			"RAW_VALUE":      &p.idx.raw_value,
			"THRESH":         &p.idx.threshold,
			"VALUE":          &p.idx.value,
			"WORST":          &p.idx.worst,
		}
		for i, f := range fs {
			pi := fieldMap[f]
			if pi == nil {
				continue
			}
			if *pi != 0 {
				return nil, fmt.Errorf("parseSMARTAttrs: duplicate header field %q in %q.", f, l)
			}
			*pi = i
		}

		if !p.idx.allSet() {
			return nil, fmt.Errorf("parseSMARTAttrs: not all fields discovered in %q", l)
		}
		return p, nil
	}

	if p.idx == nil {
		return nil, fmt.Errorf("parseSMARTAttrs: non-header line before header in SMART Attributes section: %q", l)
	}

	var err error
	var a SMARTAttribute
	var n int64

	if n, err = strconv.ParseInt(fs[0], 0, 32); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute ID %q: %v", fs[0], err)
	}
	a.ID = int32(n)

	a.Name = strings.ToLower(fs[p.idx.name])

	if n, err = strconv.ParseInt(fs[p.idx.value], 0, 32); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute current value %q: %v", fs[p.idx.value], err)
	}
	a.Value = int32(n)

	if n, err = strconv.ParseInt(fs[p.idx.worst], 0, 32); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute worst value %q: %v", fs[p.idx.worst], err)
	}
	a.Worst = int32(n)

	if n, err = strconv.ParseInt(fs[p.idx.threshold], 0, 32); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute threshold %q: %v", fs[p.idx.threshold], err)
	}
	a.Threshold = int32(n)

	rawFields := strings.SplitN(fs[p.idx.raw_value], " ", 2)
	if n, err = strconv.ParseInt(rawFields[0], 10, 64); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute raw value %q: %v", rawFields[0], err)
	}
	a.Raw.Value = n
	a.Raw.Text = fs[p.idx.raw_value]

	if n, err = strconv.ParseInt(fs[p.idx.flag], 0, 32); err != nil {
		return nil, fmt.Errorf("parseSMARTAttrs: cannot parse attribute flags %q: %v", fs[p.idx.flag], err)
	}
	a.Flags.Value = n

	if a.Flags.Value&0x0001 != 0 {
		a.Flags.Prefailure = true
		a.Flags.Text = a.Flags.Text + "P"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0x0002 != 0 {
		a.Flags.UpdatedOnline = true
		a.Flags.Text = a.Flags.Text + "O"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0x0004 != 0 {
		a.Flags.Performance = true
		a.Flags.Text = a.Flags.Text + "S"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0x0008 != 0 {
		a.Flags.ErrorRate = true
		a.Flags.Text = a.Flags.Text + "R"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0x0010 != 0 {
		a.Flags.EventCount = true
		a.Flags.Text = a.Flags.Text + "C"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0x0020 != 0 {
		a.Flags.AutoKeep = true
		a.Flags.Text = a.Flags.Text + "K"
	} else {
		a.Flags.Text = a.Flags.Text + "-"
	}

	if a.Flags.Value&0xffc0 != 0 {
		a.Flags.Text = a.Flags.Text + "+"
	} else {
		a.Flags.Text = a.Flags.Text + " "
	}

	o.ATASMARTAttributes.Table = append(o.ATASMARTAttributes.Table, &a)

	return p, nil
}

func parseSmartCtl2Prom(o *Output, l string) (lineParser, error) {
	if l == "" {
		return nil, nil
	}
	f := strings.SplitN(l, " ", 2)
	if len(f) != 2 {
		return nil, fmt.Errorf("parseSmartCtl2Prom: cannot split %q into key:value pair", l)
	}
	switch k, v := strings.ToLower(strings.TrimSpace(f[0])), strings.TrimSpace(f[1]); k {
	case "device":
		o.Device.Name = v
	case "exit_code":
		var n int64
		var err error
		if n, err = strconv.ParseInt(v, 10, 8); err != nil {
			return nil, fmt.Errorf("parseSmartCtl2Prom: cannot parse exit code %q as decimal integer: %v", v, err)
		}
		o.SmartCtl.ExitStatus = int(n)
	case "timestamp":
		var n int64
		var err error
		if n, err = strconv.ParseInt(v, 10, 64); err != nil {
			return nil, fmt.Errorf("parseSmartCtl2Prom: cannot parse timestamp %q as decimal integer: %v", v, err)
		}
		if o.LocalTime.TimeT != 0 {
			break
		}
		o.LocalTime.TimeT = n
		o.LocalTime.AscTime = time.Unix(n, 0).Format(smartCtlDate)
	}
	return lineParserFunc(parseSmartCtl2Prom), nil
}
