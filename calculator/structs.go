package calculator

import (
	"encoding/xml"
	"time"
)

type PubMarketTime struct {
	time.Time
}

func (p *PubMarketTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	t, err := time.Parse("2006-01-02T15:04Z", v)
	if err != nil {
		return err
	}
	p.Time = t
	return nil
}

type PublicationMarketDocument struct {
	XMLName                     xml.Name `xml:"Publication_MarketDocument"`
	Text                        string   `xml:",chardata"`
	Xmlns                       string   `xml:"xmlns,attr"`
	MRID                        string   `xml:"mRID"`
	RevisionNumber              string   `xml:"revisionNumber"`
	Type                        string   `xml:"type"`
	SenderMarketParticipantMRID struct {
		Text         string `xml:",chardata"`
		CodingScheme string `xml:"codingScheme,attr"`
	} `xml:"sender_MarketParticipant.mRID"`
	SenderMarketParticipantMarketRoleType string `xml:"sender_MarketParticipant.marketRole.type"`
	ReceiverMarketParticipantMRID         struct {
		Text         string `xml:",chardata"`
		CodingScheme string `xml:"codingScheme,attr"`
	} `xml:"receiver_MarketParticipant.mRID"`
	ReceiverMarketParticipantMarketRoleType string    `xml:"receiver_MarketParticipant.marketRole.type"`
	CreatedDateTime                         time.Time `xml:"createdDateTime,string"`
	PeriodTimeInterval                      struct {
		Text  string        `xml:",chardata"`
		Start PubMarketTime `xml:"start"`
		End   PubMarketTime `xml:"end"`
	} `xml:"period.timeInterval"`
	TimeSeries struct {
		Text         string `xml:",chardata"`
		MRID         string `xml:"mRID"`
		BusinessType string `xml:"businessType"`
		InDomainMRID struct {
			Text         string `xml:",chardata"`
			CodingScheme string `xml:"codingScheme,attr"`
		} `xml:"in_Domain.mRID"`
		OutDomainMRID struct {
			Text         string `xml:",chardata"`
			CodingScheme string `xml:"codingScheme,attr"`
		} `xml:"out_Domain.mRID"`
		CurrencyUnitName     string `xml:"currency_Unit.name"`
		PriceMeasureUnitName string `xml:"price_Measure_Unit.name"`
		CurveType            string `xml:"curveType"`
		Period               struct {
			Text         string `xml:",chardata"`
			TimeInterval struct {
				Text  string        `xml:",chardata"`
				Start PubMarketTime `xml:"start"`
				End   PubMarketTime `xml:"end"`
			} `xml:"timeInterval"`
			Resolution string `xml:"resolution"`
			Point      []struct {
				Text        string  `xml:",chardata"`
				Position    int     `xml:"position,string"`
				PriceAmount float64 `xml:"price.amount,string"`
			} `xml:"Point"`
		} `xml:"Period"`
	} `xml:"TimeSeries"`
}

type ExchangeRateInfo struct {
	XMLName        xml.Name `xml:"StructureSpecificData"`
	Text           string   `xml:",chardata"`
	Ss             string   `xml:"ss,attr"`
	Footer         string   `xml:"footer,attr"`
	Ns1            string   `xml:"ns1,attr"`
	Message        string   `xml:"message,attr"`
	Common         string   `xml:"common,attr"`
	Xsi            string   `xml:"xsi,attr"`
	XML            string   `xml:"xml,attr"`
	SchemaLocation string   `xml:"schemaLocation,attr"`
	Header         struct {
		Text     string `xml:",chardata"`
		ID       string `xml:"ID"`
		Test     string `xml:"Test"`
		Prepared string `xml:"Prepared"`
		Sender   struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"Sender"`
		Receiver struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
		} `xml:"Receiver"`
		Structure struct {
			Text                   string `xml:",chardata"`
			StructureID            string `xml:"structureID,attr"`
			Namespace              string `xml:"namespace,attr"`
			DimensionAtObservation string `xml:"dimensionAtObservation,attr"`
			StructureUsage         struct {
				Text string `xml:",chardata"`
				Ref  struct {
					Text     string `xml:",chardata"`
					AgencyID string `xml:"agencyID,attr"`
					ID       string `xml:"id,attr"`
					Version  string `xml:"version,attr"`
				} `xml:"Ref"`
			} `xml:"StructureUsage"`
		} `xml:"Structure"`
		DataSetAction  string `xml:"DataSetAction"`
		Extracted      string `xml:"Extracted"`
		ReportingBegin string `xml:"ReportingBegin"`
		ReportingEnd   string `xml:"ReportingEnd"`
	} `xml:"Header"`
	DataSet struct {
		Text         string `xml:",chardata"`
		DataScope    string `xml:"dataScope,attr"`
		Type         string `xml:"type,attr"`
		StructureRef string `xml:"structureRef,attr"`
		Series       struct {
			Text       string `xml:",chardata"`
			FREQ       string `xml:"FREQ,attr"`
			BASECUR    string `xml:"BASE_CUR,attr"`
			QUOTECUR   string `xml:"QUOTE_CUR,attr"`
			TENOR      string `xml:"TENOR,attr"`
			DECIMALS   string `xml:"DECIMALS,attr"`
			CALCULATED string `xml:"CALCULATED,attr"`
			UNITMULT   int    `xml:"UNIT_MULT,attr,string"`
			COLLECTION string `xml:"COLLECTION,attr"`
			Obs        struct {
				Text       string  `xml:",chardata"`
				TIMEPERIOD string  `xml:"TIME_PERIOD,attr"`
				OBSVALUE   float64 `xml:"OBS_VALUE,attr,string"`
			} `xml:"Obs"`
		} `xml:"Series"`
	} `xml:"DataSet"`
}
