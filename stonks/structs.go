package stonks

import "encoding/json"

//StockSymbol -- from datahub file
type StockSymbol struct {
	ACTSymbol   string `json:"ACT Symbol"`
	CompanyName string `json:"Company Name"`
}

//Thing - a generic reddit datatype
type Thing struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	Data RedditData
}

//RedditData - data container class
type RedditData struct {
	Data struct {
		CommentData
		LinkData
		Votable
		Created
		MoreData
		ID string `json:"id"`
	} `json:"data"`

	Kind string `json:"kind"`
}

//Votable implementation
type Votable struct {
	Ups   int64 `json:"ups"`
	Downs int64 `json:"downs"`
}

//Created implementation
type Created struct {
	CreatedUTC float64 `json:"created_utc"`
	ID         string  `json:"id"`
}

//LinkListing - a listing of links
type LinkListing struct {
	Before   string `json:"before"`
	After    string `json:"after"`
	ModHash  string `json:"modhash"`
	Children []Link `json:"children"`
}

//DataListing - a listing of data
type DataListing struct {
	Data ListingData `json:"data"`
	Kind string      `json:"kind"`
}

//ListingData - data for listing
type ListingData struct {
	Before   string       `json:"before"`
	After    string       `json:"after"`
	ModHash  string       `json:"modhash"`
	Children []RedditData `json:"children"`
}

//Comment - a reddit comment
type Comment struct {
	Data CommentData `json:"data"`
	Kind string      `json:"kind"`
	Created
	Votable
	ID string `json:"id"`
}

//CommentData - data in a comment struct
type CommentData struct {
	ApprovedBy string `json:"approved_by"`
	Author     string `json:"author"`
	Body       string `json:"body"`
	ParentID   string `json:"parent_id"`
	LinkID     string `json:"link_id"`
	LinkTitle  string `json:"link_title"`
	//	Replies    []Thing `json:"replies"`
}

//LinkData - data in a link struct
type LinkData struct {
	Media    interface{} `json:"media"`
	SelfText string      `json:"selftext"`
}

//Link - a struct containing a link
type Link struct {
	Data LinkData `json:"data"`
	Kind string   `json:"kind"`
	Created
	Votable
	ID string `json:"id"`
}

//MoreData - additional comment IDs
type MoreData struct {
	Children []string `json:"children"`
}

//UnmarshalJSON -- implement JSON Unmarshaler
func (l *Link) UnmarshalJSON(b []byte) (err error) {
	return json.Unmarshal(b, &l)
}

//JSONResponse - a dataListing slice
type JSONResponse []DataListing

//GetAllComments - get all comments contained in a DataListing
func (d DataListing) GetAllComments() (c []Comment) {
	c = make([]Comment, 0, len(d.Data.Children))
	for _, v := range d.Data.Children {
		if v.Kind == KComment {
			x := Comment{Data: v.Data.CommentData, Created: v.Data.Created, Votable: v.Data.Votable, ID: v.Data.ID}
			c = append(c, x)

		}
	}

	return
}

//GetAllLinks - get all links contained in a DataListing
func (d DataListing) GetAllLinks() (l []Link) {
	l = make([]Link, 0, 16)
	for _, v := range d.Data.Children {
		if v.Kind == KLink {
			x := Link{Data: v.Data.LinkData, Created: v.Data.Created, Votable: v.Data.Votable, ID: v.Data.ID}
			l = append(l, x)

		}
	}

	return
}

//GetMore - get a slice of More
func (d DataListing) GetMore() (m []MoreData) {
	for _, v := range d.Data.Children {
		if v.Kind == KMore {
			m = append(m, v.Data.MoreData)
		}
	}
	return
}
