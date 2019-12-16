package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"badrpc/smartctl2prom/smartctldata"
)

func main() {
	if len(os.Args) != 1 {
		log.Fatal("usage: smartctl2prom filename")
	}

	// Standard registry in prometheus module adds a number of internal
	// process metrics which result in duplicate metrics if more than one
	// text file exprter does this. An empty registry will not have those.
	reg := prometheus.NewRegistry()

	readTime := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_read_time",
		Help: "Time when SMART data were read from device.",
	}, []string{
		"device_name",
		"device_type",
		"device_model_family",
		"device_model_name",
		"device_serial_number",
		"smartctl_exit_status",
	})
	reg.MustRegister(readTime)

	deviceIdLabels := []string{"device_name", "device_serial_number"}
	capacityBlocks := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_user_capacity_blocks",
	}, deviceIdLabels)
	reg.MustRegister(capacityBlocks)
	capacityBytes := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_user_capacity_bytes",
	}, deviceIdLabels)
	reg.MustRegister(capacityBytes)
	logicalBlockSize := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_logical_block_size_bytes",
	}, deviceIdLabels)
	reg.MustRegister(logicalBlockSize)
	physicalBlockSize := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_physical_block_size_bytes",
	}, deviceIdLabels)
	reg.MustRegister(physicalBlockSize)
	interfaceSpeed := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_interface_speed_bps",
	}, deviceIdLabels)
	reg.MustRegister(interfaceSpeed)
	selfAssessmentPassed := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_overall_health_self_assessment_passed",
	}, deviceIdLabels)
	reg.MustRegister(selfAssessmentPassed)
	powerOnHours := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_power_on_hours",
	}, deviceIdLabels)
	reg.MustRegister(powerOnHours)
	powerCycles := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_power_cycles_total",
	}, deviceIdLabels)
	reg.MustRegister(powerCycles)
	temperature := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_temperature_celsius",
	}, deviceIdLabels)
	reg.MustRegister(temperature)

	attributeLabels := []string{"id", "name", "prefailure", "device_name", "device_serial_number"}
	attributeValue := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_ata_attribute_value",
	}, attributeLabels)
	reg.MustRegister(attributeValue)
	attributeWorst := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_ata_attribute_worst",
	}, attributeLabels)
	reg.MustRegister(attributeWorst)
	attributeThresh := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_ata_attribute_thresh",
	}, attributeLabels)
	reg.MustRegister(attributeThresh)
	attributeRaw := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "smart_device_ata_attribute_raw_value",
	}, attributeLabels)
	reg.MustRegister(attributeRaw)
	// TODO(badrpc): have not figured out how to create _min and _max only
	// when they are provided in smartctl output.
	//
	// attributeRawMin := prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Name: "smart_device_ata_attribute_raw_value_min",
	// }, attributeLabels)
	// reg.MustRegister(attributeRawMin)
	// attributeRawMax := prometheus.NewGaugeVec(prometheus.GaugeOpts{
	// 	Name: "smart_device_ata_attribute_raw_value_max",
	// }, attributeLabels)
	// reg.MustRegister(attributeRawMax)

	for oe := range smartctldata.DecodeJSON(os.Stdin) {
		if oe.Err != nil {
			log.Print(oe.Err)
			continue
		}
		o := oe.O

		readTime.With(prometheus.Labels{
			"device_name":          o.Device.Name,
			"device_type":          o.Device.Type,
			"device_model_family":  o.ModelFamily,
			"device_model_name":    o.ModelName,
			"device_serial_number": o.SerialNumber,
			"smartctl_exit_status": strconv.Itoa(o.SmartCtl.ExitStatus),
		}).Set(float64(o.LocalTime.TimeT))
		capacityBlocks.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.UserCapacity.Blocks))
		capacityBytes.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.UserCapacity.Bytes))
		logicalBlockSize.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.LogicalBlockSize))
		physicalBlockSize.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.PhysicalBlockSize))
		interfaceSpeed.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.InterfaceSpeed.Current.UnitsPerSecond * o.InterfaceSpeed.Current.BitsPerUnit))
		var selfAssessmentPassedVal = 0.0
		if o.SMARTStatus.Passed {
			selfAssessmentPassedVal = 1.0
		}
		selfAssessmentPassed.WithLabelValues(o.Device.Name, o.SerialNumber).Set(selfAssessmentPassedVal)
		powerOnHours.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.PowerOnTime.Hours))
		powerCycles.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.PowerCycleCount))
		temperature.WithLabelValues(o.Device.Name, o.SerialNumber).Set(float64(o.Temperature.Current))

		for _, a := range o.ATASMARTAttributes.Table {
			preFailure := "no"
			if a.Flags.Prefailure {
				preFailure = "yes"
			}
			attrLabels := prometheus.Labels{
				"id":                   strconv.Itoa(int(a.ID)),
				"name":                 strings.ToLower(a.Name),
				"prefailure":           preFailure,
				"device_name":          o.Device.Name,
				"device_serial_number": o.SerialNumber,
			}
			attributeValue.With(attrLabels).Set(float64(a.Value))
			attributeWorst.With(attrLabels).Set(float64(a.Worst))
			attributeThresh.With(attrLabels).Set(float64(a.Threshold))
			readRawValue(a, attributeRaw.With(attrLabels) /*, attributeRawMin.With(attrLabels), attributeRawMax.With(attrLabels)*/)
		}
	}

	if err := prometheus.WriteToTextfile(os.Args[1], reg); err != nil {
		log.Fatal(err)
	}
}

func readRawValue(a *smartctldata.SMARTAttribute, raw /*, min, max */ prometheus.Gauge) {
	rawValue := a.Raw.Value
	switch a.ID {
	case 194: // Temperature_Celsius
		switch {
		case rawValue <= 0x00ffff:
			raw.Set(float64(rawValue))
		case rawValue <= 0xffffffffffff && rawValue > 0x0000ffffffff:
			raw.Set(float64(rawValue & 0x00000000ffff))
			// min.Set(float64((rawValue & 0x0000ffff0000) >> 16))
			// max.Set(float64((rawValue & 0xffff00000000) >> 32))
		default:
			raw.Set(float64(rawValue))
		}
	default:
		raw.Set(float64(rawValue))
	}
}
