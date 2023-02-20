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

type AcknowledgementMarketDocument struct {
	MRID                                    string `json:"mRID"`
	CreatedDateTime                         string `json:"createdDateTime"`
	SenderMarketParticipantMarketRoleType   string `json:"sender_MarketParticipant.marketRole.type"`
	ReceiverMarketParticipantMarketRoleType string `json:"receiver_MarketParticipant.marketRole.type"`
	ReceivedMarketDocumentCreatedDateTime   string `json:"received_MarketDocument.createdDateTime"`
	Reason                                  struct {
		Code string `json:"code"`
		Text string `json:"text"`
	} `json:"Reason"`
	Xmlns string `json:"_xmlns"`
}
