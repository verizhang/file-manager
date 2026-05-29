package model

type VirusScanStatus string

const (
	VirusScanStatusPending  VirusScanStatus = "PENDING"
	VirusScanStatusScaning  VirusScanStatus = "SCANING"
	VirusScanStatusClean    VirusScanStatus = "CLEAN"
	VirusScanStatusInfected VirusScanStatus = "INFECTED"
	VirusScanStatusFailed   VirusScanStatus = "FAILED"
)
