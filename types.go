package booking

import "strings"

type Anchorage struct {
	Id                        interface{}   `json:"id"`
	ApplyNumber               interface{}   `json:"applyNumber"`
	AcceptHsjg                interface{}   `json:"acceptHsjg"`
	ShipNameCh                interface{}   `json:"shipNameCh"`
	ShipNameEn                interface{}   `json:"shipNameEn"`
	ShipFirstRegNo            interface{}   `json:"shipFirstRegNo"`
	ShipId                    interface{}   `json:"shipId"`
	ShipRegNo                 interface{}   `json:"shipRegNo"`
	Cbbh                      interface{}   `json:"cbbh"`
	Mmsi                      interface{}   `json:"mmsi"`
	Imo                       interface{}   `json:"imo"`
	Callsign                  interface{}   `json:"callsign"`
	IsChinaShip               interface{}   `json:"isChinaShip"`
	ShipNationality           interface{}   `json:"shipNationality"`
	ShipType                  interface{}   `json:"shipType"`
	ShipNativePort            interface{}   `json:"shipNativePort"`
	SeaOrRiver                interface{}   `json:"seaOrRiver"`
	ShipLength                interface{}   `json:"shipLength"`
	ShipBreadth               interface{}   `json:"shipBreadth"`
	TotalTonnage              interface{}   `json:"totalTonnage"`
	CleanTonnage              interface{}   `json:"cleanTonnage"`
	ShipDwt                   interface{}   `json:"shipDwt"`
	MachinePower              interface{}   `json:"machinePower"`
	BuiltDate                 interface{}   `json:"builtDate"`
	ShipOwner                 interface{}   `json:"shipOwner"`
	ShipManagerPerson         interface{}   `json:"shipManagerPerson"`
	SatellitePhone            interface{}   `json:"satellitePhone"`
	LoadCargo                 interface{}   `json:"loadCargo"`
	EventStatus               interface{}   `json:"eventStatus"`
	PredictAnchorGround       interface{}   `json:"predictAnchorGround"`
	PredictAnchorPosition     interface{}   `json:"predictAnchorPosition"`
	PredictAnchorTime         interface{}   `json:"predictAnchorTime"`
	PredictMoveAnchorTime     interface{}   `json:"predictMoveAnchorTime"`
	ArrangeAnchorTime         interface{}   `json:"arrangeAnchorTime"`
	ArrangeMoveAnchorTime     interface{}   `json:"arrangeMoveAnchorTime"`
	ActualAnchorTime          interface{}   `json:"actualAnchorTime"`
	ActualMoveAnchorTime      interface{}   `json:"actualMoveAnchorTime"`
	StopReason                interface{}   `json:"stopReason"`
	OtherReason               interface{}   `json:"otherReason"`
	NextPortBerth             interface{}   `json:"nextPortBerth"`
	CargoLoading              interface{}   `json:"cargoLoading"`
	MaxDraft                  interface{}   `json:"maxDraft"`
	RealDraft                 interface{}   `json:"realDraft"`
	WetherDangerousGoods      interface{}   `json:"wetherDangerousGoods"`
	ForeignCrew               interface{}   `json:"foreignCrew"`
	ChineseCrew               interface{}   `json:"chineseCrew"`
	PilotInfoJsons            interface{}   `json:"pilotInfoJsons"`
	TowBoatInfoJsons          interface{}   `json:"towBoatInfoJsons"`
	WhetherAgainApply         interface{}   `json:"whetherAgainApply"`
	IsNewData                 interface{}   `json:"isNewData"`
	WhetherHandleLimit        interface{}   `json:"whetherHandleLimit"`
	ProxyInfoJson             interface{}   `json:"proxyInfoJson"`
	DropAnchorLocation        interface{}   `json:"dropAnchorLocation"`
	PublishTime               interface{}   `json:"publishTime"`
	PublishUser               interface{}   `json:"publishUser"`
	CheckOpinion              interface{}   `json:"checkOpinion"`
	OperateTime               interface{}   `json:"operateTime"`
	Remark                    interface{}   `json:"remark"`
	WhetherSpecial            interface{}   `json:"whetherSpecial"`
	SpecialContent            interface{}   `json:"specialContent"`
	IsSubmit                  interface{}   `json:"isSubmit"`
	ApplyItem                 interface{}   `json:"applyItem"`
	ApplyObject               interface{}   `json:"applyObject"`
	DealPerson                interface{}   `json:"dealPerson"`
	ApplyType                 interface{}   `json:"applyType"`
	ContactPhone              interface{}   `json:"contactPhone"`
	ApplyDataSource           interface{}   `json:"applyDataSource"`
	ApplyTime                 interface{}   `json:"applyTime"`
	PilotInfoJsonList         []interface{} `json:"pilotInfoJsonList"`
	TowBoatInfoJsonList       []interface{} `json:"towBoatInfoJsonList"`
	ProxyInfoJsonDTO          interface{}   `json:"proxyInfoJsonDTO"`
	FileList                  []File        `json:"fileList"`
	AnchorType                interface{}   `json:"anchorType"`
	PredictAnchorGroundName   interface{}   `json:"predictAnchorGroundName"`
	PredictAnchorPositionName interface{}   `json:"predictAnchorPositionName"`
	ArrangeTimeLen            interface{}   `json:"arrangeTimeLen"`
	List                      interface{}   `json:"list"`
	EventStatusValue          interface{}   `json:"eventStatusValue"`
	EventStatusName           interface{}   `json:"eventStatusName"`
	CreateDate                interface{}   `json:"createDate"`
	UpdateDate                interface{}   `json:"updateDate"`
	UserWithDrawRemark        interface{}   `json:"userWithDrawRemark"`
	ShipNation                interface{}   `json:"shipNation"`
	TimeNum                   interface{}   `json:"timeNum"`
	PortArea                  interface{}   `json:"portArea"`
	PortAreaName              interface{}   `json:"portAreaName"`
	InOrOut                   interface{}   `json:"inOrOut"`
	PortStatus                interface{}   `json:"portStatus"`
	PortStatusName            interface{}   `json:"portStatusName"`
	NewShipFlag               interface{}   `json:"newShipFlag"`
	AlarmNum                  interface{}   `json:"alarmNum"`
	HandleNum                 interface{}   `json:"handleNum"`
	WhetherAutoPublic         interface{}   `json:"whetherAutoPublic"`
	DownUploadfileList        []File        `json:"downUploadfileList"`
	StopReasonList            []interface{} `json:"stopReasonList"`
	IsAnchGroundLimit         interface{}   `json:"isAnchGroundLimit"`
}

type File struct {
	Id         interface{} `json:"id"`
	BusinessId interface{} `json:"businessId"`
	FileLabel  interface{} `json:"fileLabel"`
	FileName   interface{} `json:"fileName"`
	FileType   interface{} `json:"fileType"`
	FileUrl    interface{} `json:"fileUrl"`
	FileSize   interface{} `json:"fileSize"`
	BucketName interface{} `json:"bucketName"`
	Remark     interface{} `json:"remark"`
}

func (c *Anchorage) normalize() {
	c.IsSubmit = 1
	// TODO fetch from getAnchorGroundList
	c.IsAnchGroundLimit = "1"
	c.DownUploadfileList = []File{}
	if len(c.FileList) != 0 {
		c.DownUploadfileList = c.FileList
	}
	c.StopReasonList = []interface{}{}
	if c.StopReason != nil {
		reasons := strings.Split(c.StopReason.(string), ",")
		for _, r := range reasons {
			c.StopReasonList = append(c.StopReasonList, r)
		}
	}
	c.DownUploadfileList = c.FileList
}
