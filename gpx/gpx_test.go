package gpx

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	TIME_FORMAT    = "2006-01-02T15:04:05Z"
	TEST_FILES_DIR = "../test_files"
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func cca(x, y float64) bool {
	return math.Abs(x-y) < 0.001
}

func assertEquals(t *testing.T, var1 interface{}, var2 interface{}) {
	if var1 != var2 {
		fmt.Println(var1, "not equals to", var2)
		t.Error("Not equals")
	}
}

func assertTrue(t *testing.T, message string, expr bool) {
	if !expr {
		t.Error(message)
	}
}

func assertLinesEquals(t *testing.T, string1, string2 string) {
	lines1 := strings.Split(string1, "\n")
	lines2 := strings.Split(string2, "\n")
	for i := 0; i < min(len(lines1), len(lines2)); i++ {
		line1 := strings.Trim(lines1[i], " \n\r\t")
		line2 := strings.Trim(lines2[i], " \n\r\t")
		if line1 != line2 {
			t.Error("Line (#", i, ") different:", line1, "\nand:", line2)
			break
		}
	}
	if len(lines1) != len(lines2) {
		fmt.Println("String1:", string1)
		fmt.Println("String2:", string2)
		t.Error("String have a different number of lines", len(lines1), "and", len(lines2))
		return
	}
}

func assertNil(t *testing.T, var1 interface{}) {
	if var1 != nil {
		fmt.Println(var1)
		t.Error("nil!")
	}
}

func assertNotNil(t *testing.T, var1 interface{}) {
	if var1 == nil {
		fmt.Println(var1)
		t.Error("nil!")
	}
}

func loadTestGPXs() []string {
	gpxes := make([]string, 0)
	dirs, _ := ioutil.ReadDir(TEST_FILES_DIR)
	for _, fileInfo := range dirs {
		if strings.HasSuffix(fileInfo.Name(), ".gpx") {
			gpxes = append(gpxes, fmt.Sprintf("%s/%s", TEST_FILES_DIR, fileInfo.Name()))
		}
	}
	if len(gpxes) == 0 {
		panic("No GPX files found")
	}
	return gpxes
}

func getMinDistanceBetweenTrackPoints(g GPX) float64 {
	result := -1.0
	for _, track := range g.Tracks {
		for _, segment := range track.Segments {
			if len(segment.Points) > 1 {
				for pointNo, point := range segment.Points {
					if pointNo > 0 {
						previousPoint := segment.Points[pointNo-1]
						distance := point.Distance3D(&previousPoint.Point)
						//fmt.Printf("distance=%f\n", distance)
						if result < 0.0 || distance < result {
							result = distance
						}
					}
				}
			}
		}
	}
	if result < 0.0 {
		return 0.0
	}
	return result
}

func TestParseGPXTimes(t *testing.T) {
	datetimes := []string{
		"2013-01-02T12:07:08Z",
		"2013-01-02 12:07:08Z",
		"2013-01-02T12:07:08",
		"2013-01-02T12:07:08.034Z",
		"2013-01-02 12:07:08.045Z",
		"2013-01-02T12:07:08.123",
	}
	for _, value := range datetimes {
		fmt.Println("datetime:", value)
		parsedTime, err := parseGPXTime(value)
		fmt.Println(parsedTime)
		assertNil(t, err)
		assertNotNil(t, parsedTime)
		assertEquals(t, parsedTime.Year(), 2013)
		assertEquals(t, parsedTime.Month(), time.January)
		assertEquals(t, parsedTime.Day(), 2)
		assertEquals(t, parsedTime.Hour(), 12)
		assertEquals(t, parsedTime.Minute(), 7)
		assertEquals(t, parsedTime.Second(), 8)
	}
}

func testDetectVersion(t *testing.T, fileName, expectedVersion string) {
	f, err := os.Open(fileName)
	fmt.Println("err=", err)
	contents, _ := ioutil.ReadAll(f)
	version, err := guessGPXVersion(contents)
	fmt.Println("Version=", version)
	if err != nil {
		t.Error("Can't detect 1.1 GPX, error=" + err.Error())
	}
	if version != expectedVersion {
		t.Error("Can't detect 1.1 GPX")
	}
}

func TestDetect11GPXVersion(t *testing.T) {
	testDetectVersion(t, "../test_files/gpx1.1_with_all_fields.gpx", "1.1")
}

func TestDetect10GPXVersion(t *testing.T) {
	testDetectVersion(t, "../test_files/gpx1.0_with_all_fields.gpx", "1.0")
}

func TestParseAndReparseGPX11(t *testing.T) {
	gpxDocuments := []*GPX{}

	{
		gpxDoc, err := ParseFile("../test_files/gpx1.1_with_all_fields.gpx")
		if err != nil || gpxDoc == nil {
			t.Error("Error parsing:" + err.Error())
		}
		gpxDocuments = append(gpxDocuments, gpxDoc)
		assertEquals(t, gpxDoc.Version, "1.1")

		// Test after reparsing
		xml, err := gpxDoc.ToXml(ToXmlParams{Version: "1.1", Indent: true})
		//fmt.Println(string(xml))
		if err != nil {
			t.Error("Error serializing to XML:" + err.Error())
		}
		gpxDoc2, err := ParseBytes(xml)
		assertEquals(t, gpxDoc2.Version, "1.1")
		if err != nil {
			t.Error("Error parsing XML:" + err.Error())
		}
		gpxDocuments = append(gpxDocuments, gpxDoc2)

		// TODO: ToString 1.0 and check again
	}

	for i := 1; i < len(gpxDocuments); i++ {
		fmt.Println("Testing gpx doc #", i)

		gpxDoc := gpxDocuments[i]

		executeSample11GpxAsserts(t, gpxDoc)

		// Tests after reparsing as 1.0
	}
}

func executeSample11GpxAsserts(t *testing.T, gpxDoc *GPX) {
	assertEquals(t, gpxDoc.Version, "1.1")
	assertEquals(t, gpxDoc.Creator, "...")
	assertEquals(t, gpxDoc.Name, "example name")
	assertEquals(t, gpxDoc.AuthorName, "author name")
	assertEquals(t, gpxDoc.AuthorEmail, "aaa@bbb.com")
	assertEquals(t, gpxDoc.Description, "example description")
	assertEquals(t, gpxDoc.AuthorLink, "http://link")
	assertEquals(t, gpxDoc.AuthorLinkText, "link text")
	assertEquals(t, gpxDoc.AuthorLinkType, "link type")
	assertEquals(t, gpxDoc.Copyright, "gpxauth")
	assertEquals(t, gpxDoc.CopyrightYear, "2013")
	assertEquals(t, gpxDoc.CopyrightLicense, "lic")
	assertEquals(t, gpxDoc.Link, "http://link2")
	assertEquals(t, gpxDoc.LinkText, "link text2")
	assertEquals(t, gpxDoc.LinkType, "link type2")
	assertEquals(t, gpxDoc.Time.Format(TIME_FORMAT), time.Date(2013, time.January, 01, 12, 0, 0, 0, time.UTC).Format(TIME_FORMAT))
	assertEquals(t, gpxDoc.Keywords, "example keywords")

	// Waypoints:
	assertEquals(t, len(gpxDoc.Waypoints), 2)
	assertEquals(t, gpxDoc.Waypoints[0].Latitude, 12.3)
	assertEquals(t, gpxDoc.Waypoints[0].Longitude, 45.6)
	assertEquals(t, gpxDoc.Waypoints[0].Elevation.Value(), 75.1)
	assertEquals(t, gpxDoc.Waypoints[0].Timestamp.Format(TIME_FORMAT), "2013-01-02T02:03:00Z")
	assertEquals(t, gpxDoc.Waypoints[0].MagneticVariation, "1.1")
	assertEquals(t, gpxDoc.Waypoints[0].GeoidHeight, "2.0")
	assertEquals(t, gpxDoc.Waypoints[0].Name, "example name")
	assertEquals(t, gpxDoc.Waypoints[0].Comment, "example cmt")
	assertEquals(t, gpxDoc.Waypoints[0].Description, "example desc")
	assertEquals(t, gpxDoc.Waypoints[0].Source, "example src")
	// TODO
	// Links       []GpxLink
	assertEquals(t, gpxDoc.Waypoints[0].Symbol, "example sym")
	assertEquals(t, gpxDoc.Waypoints[0].Type, "example type")
	assertEquals(t, gpxDoc.Waypoints[0].TypeOfGpsFix, "2d")
	assertEquals(t, gpxDoc.Waypoints[0].Satellites.Value(), 5)
	assertEquals(t, gpxDoc.Waypoints[0].HorizontalDilution.Value(), 6.0)
	assertEquals(t, gpxDoc.Waypoints[0].VerticalDilution.Value(), 7.0)
	assertEquals(t, gpxDoc.Waypoints[0].PositionalDilution.Value(), 8.0)
	assertEquals(t, gpxDoc.Waypoints[0].AgeOfDGpsData.Value(), 9.0)
	assertEquals(t, gpxDoc.Waypoints[0].DGpsId.Value(), 45)
	// TODO: Extensions

	assertEquals(t, gpxDoc.Waypoints[1].Latitude, 13.4)
	assertEquals(t, gpxDoc.Waypoints[1].Longitude, 46.7)

	// Routes:
	assertEquals(t, len(gpxDoc.Routes), 2)
	assertEquals(t, gpxDoc.Routes[0].Name, "example name")
	assertEquals(t, gpxDoc.Routes[0].Comment, "example cmt")
	assertEquals(t, gpxDoc.Routes[0].Description, "example desc")
	assertEquals(t, gpxDoc.Routes[0].Source, "example src")
	assertEquals(t, gpxDoc.Routes[0].Number, 7)
	assertEquals(t, gpxDoc.Routes[0].Type, "rte type")
	assertEquals(t, len(gpxDoc.Routes[0].Points), 3)
	// TODO: Link
	// TODO: Points
	assertEquals(t, gpxDoc.Routes[0].Points[0].Elevation.Value(), 75.1)
	fmt.Println("t=", gpxDoc.Routes[0].Points[0].Timestamp)
	assertEquals(t, gpxDoc.Routes[0].Points[0].Timestamp.Format(TIME_FORMAT), "2013-01-02T02:03:03Z")
	assertEquals(t, gpxDoc.Routes[0].Points[0].MagneticVariation, "1.2")
	assertEquals(t, gpxDoc.Routes[0].Points[0].GeoidHeight, "2.1")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Name, "example name r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Comment, "example cmt r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Description, "example desc r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Source, "example src r")
	// TODO
	//assertEquals(t, gpxDoc.Routes[0].Points[0].Link, "http://linkrtept")
	//assertEquals(t, gpxDoc.Routes[0].Points[0].Text, "rtept link")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Type, "example type r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Symbol, "example sym r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Type, "example type r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].TypeOfGpsFix, "3d")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Satellites.Value(), 6)
	assertEquals(t, gpxDoc.Routes[0].Points[0].HorizontalDilution.Value(), 7.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].VerticalDilution.Value(), 8.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].PositionalDilution.Value(), 9.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].AgeOfDGpsData.Value(), 10.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].DGpsId.Value(), 99)
	// TODO: Extensions

	assertEquals(t, gpxDoc.Routes[1].Name, "second route")
	assertEquals(t, gpxDoc.Routes[1].Description, "example desc 2")
	assertEquals(t, len(gpxDoc.Routes[1].Points), 2)

	// Tracks:
	assertEquals(t, len(gpxDoc.Tracks), 2)
	assertEquals(t, gpxDoc.Tracks[0].Name, "example name t")
	assertEquals(t, gpxDoc.Tracks[0].Comment, "example cmt t")
	assertEquals(t, gpxDoc.Tracks[0].Description, "example desc t")
	assertEquals(t, gpxDoc.Tracks[0].Source, "example src t")
	assertEquals(t, gpxDoc.Tracks[0].Number, 1)
	assertEquals(t, gpxDoc.Tracks[0].Type, "t")
	// TODO link

	assertEquals(t, len(gpxDoc.Tracks[0].Segments), 2)

	assertEquals(t, len(gpxDoc.Tracks[0].Segments[0].Points), 1)
	assertEquals(t, len(gpxDoc.Tracks[0].Segments[1].Points), 0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Elevation.Value(), 11.1)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Timestamp.Format(TIME_FORMAT), "2013-01-01T12:00:04Z")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].MagneticVariation, "12")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].GeoidHeight, "13")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Name, "example name t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Comment, "example cmt t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Description, "example desc t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Source, "example src t")
	// TODO link
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Symbol, "example sym t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Type, "example type t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].TypeOfGpsFix, "3d")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Satellites.Value(), 100)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].HorizontalDilution.Value(), 101.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].VerticalDilution.Value(), 102.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].PositionalDilution.Value(), 103.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].AgeOfDGpsData.Value(), 104.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].DGpsId.Value(), 99)
	// TODO extensions
}

func TestParseAndReparseGPX10(t *testing.T) {
	gpxDocuments := []*GPX{}

	{
		gpxDoc, err := ParseFile("../test_files/gpx1.0_with_all_fields.gpx")
		if err != nil || gpxDoc == nil {
			t.Error("Error parsing:" + err.Error())
		}
		gpxDocuments = append(gpxDocuments, gpxDoc)
		assertEquals(t, gpxDoc.Version, "1.0")

		// Test after reparsing
		xml, err := gpxDoc.ToXml(ToXmlParams{Version: "1.0", Indent: true})
		//fmt.Println(string(xml))
		if err != nil {
			t.Error("Error serializing to XML:" + err.Error())
		}
		gpxDoc2, err := ParseBytes(xml)
		assertEquals(t, gpxDoc2.Version, "1.0")
		if err != nil {
			t.Error("Error parsing XML:" + err.Error())
		}
		gpxDocuments = append(gpxDocuments, gpxDoc2)

		// TODO: ToString 1.0 and check again
	}

	for i := 1; i < len(gpxDocuments); i++ {
		fmt.Println("Testing gpx doc #", i)

		gpxDoc := gpxDocuments[i]

		executeSample10GpxAsserts(t, gpxDoc)

		// Tests after reparsing as 1.0
	}
}

func executeSample10GpxAsserts(t *testing.T, gpxDoc *GPX) {
	assertEquals(t, gpxDoc.Version, "1.0")
	assertEquals(t, gpxDoc.Creator, "...")
	assertEquals(t, gpxDoc.Name, "example name")
	assertEquals(t, gpxDoc.AuthorName, "example author")
	assertEquals(t, gpxDoc.AuthorEmail, "example@email.com")
	assertEquals(t, gpxDoc.Description, "example description")
	assertEquals(t, gpxDoc.AuthorLink, "")
	assertEquals(t, gpxDoc.AuthorLinkText, "")
	assertEquals(t, gpxDoc.AuthorLinkType, "")
	assertEquals(t, gpxDoc.Copyright, "")
	assertEquals(t, gpxDoc.CopyrightYear, "")
	assertEquals(t, gpxDoc.CopyrightLicense, "")
	assertEquals(t, gpxDoc.Link, "http://example.url")
	assertEquals(t, gpxDoc.LinkText, "example urlname")
	assertEquals(t, gpxDoc.LinkType, "")
	assertEquals(t, gpxDoc.Time.Format(TIME_FORMAT), time.Date(2013, time.January, 01, 12, 0, 0, 0, time.UTC).Format(TIME_FORMAT))
	assertEquals(t, gpxDoc.Keywords, "example keywords")

	// TODO: Bounds (here and in 1.1)

	// Waypoints:
	assertEquals(t, len(gpxDoc.Waypoints), 2)
	assertEquals(t, gpxDoc.Waypoints[0].Latitude, 12.3)
	assertEquals(t, gpxDoc.Waypoints[0].Longitude, 45.6)
	assertEquals(t, gpxDoc.Waypoints[0].Elevation.Value(), 75.1)
	assertEquals(t, gpxDoc.Waypoints[0].Timestamp.Format(TIME_FORMAT), "2013-01-02T02:03:00Z")
	assertEquals(t, gpxDoc.Waypoints[0].MagneticVariation, "1.1")
	assertEquals(t, gpxDoc.Waypoints[0].GeoidHeight, "2.0")
	assertEquals(t, gpxDoc.Waypoints[0].Name, "example name")
	assertEquals(t, gpxDoc.Waypoints[0].Comment, "example cmt")
	assertEquals(t, gpxDoc.Waypoints[0].Description, "example desc")
	assertEquals(t, gpxDoc.Waypoints[0].Source, "example src")
	// TODO
	// Links       []GpxLink
	assertEquals(t, gpxDoc.Waypoints[0].Symbol, "example sym")
	assertEquals(t, gpxDoc.Waypoints[0].Type, "example type")
	assertEquals(t, gpxDoc.Waypoints[0].TypeOfGpsFix, "2d")
	assertEquals(t, gpxDoc.Waypoints[0].Satellites.Value(), 5)
	assertEquals(t, gpxDoc.Waypoints[0].HorizontalDilution.Value(), 6.0)
	assertEquals(t, gpxDoc.Waypoints[0].VerticalDilution.Value(), 7.0)
	assertEquals(t, gpxDoc.Waypoints[0].PositionalDilution.Value(), 8.0)
	assertEquals(t, gpxDoc.Waypoints[0].AgeOfDGpsData.Value(), 9.0)
	assertEquals(t, gpxDoc.Waypoints[0].DGpsId.Value(), 45)
	// TODO: Extensions

	assertEquals(t, gpxDoc.Waypoints[1].Latitude, 13.4)
	assertEquals(t, gpxDoc.Waypoints[1].Longitude, 46.7)

	// Routes:
	assertEquals(t, len(gpxDoc.Routes), 2)
	assertEquals(t, gpxDoc.Routes[0].Name, "example name")
	assertEquals(t, gpxDoc.Routes[0].Comment, "example cmt")
	assertEquals(t, gpxDoc.Routes[0].Description, "example desc")
	assertEquals(t, gpxDoc.Routes[0].Source, "example src")
	assertEquals(t, gpxDoc.Routes[0].Number, 7)
	assertEquals(t, gpxDoc.Routes[0].Type, "")
	assertEquals(t, len(gpxDoc.Routes[0].Points), 3)
	// TODO: Link
	// TODO: Points
	assertEquals(t, gpxDoc.Routes[0].Points[0].Elevation.Value(), 75.1)
	fmt.Println("t=", gpxDoc.Routes[0].Points[0].Timestamp)
	assertEquals(t, gpxDoc.Routes[0].Points[0].Timestamp.Format(TIME_FORMAT), "2013-01-02T02:03:03Z")
	assertEquals(t, gpxDoc.Routes[0].Points[0].MagneticVariation, "1.2")
	assertEquals(t, gpxDoc.Routes[0].Points[0].GeoidHeight, "2.1")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Name, "example name r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Comment, "example cmt r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Description, "example desc r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Source, "example src r")
	// TODO link
	//assertEquals(t, gpxDoc.Routes[0].Points[0].Link, "http://linkrtept")
	//assertEquals(t, gpxDoc.Routes[0].Points[0].Text, "rtept link")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Type, "example type r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Symbol, "example sym r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Type, "example type r")
	assertEquals(t, gpxDoc.Routes[0].Points[0].TypeOfGpsFix, "3d")
	assertEquals(t, gpxDoc.Routes[0].Points[0].Satellites.Value(), 6)
	assertEquals(t, gpxDoc.Routes[0].Points[0].HorizontalDilution.Value(), 7.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].VerticalDilution.Value(), 8.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].PositionalDilution.Value(), 9.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].AgeOfDGpsData.Value(), 10.0)
	assertEquals(t, gpxDoc.Routes[0].Points[0].DGpsId.Value(), 99)
	// TODO: Extensions

	assertEquals(t, gpxDoc.Routes[1].Name, "second route")
	assertEquals(t, gpxDoc.Routes[1].Description, "example desc 2")
	assertEquals(t, len(gpxDoc.Routes[1].Points), 2)

	// Tracks:
	assertEquals(t, len(gpxDoc.Tracks), 2)
	assertEquals(t, gpxDoc.Tracks[0].Name, "example name t")
	assertEquals(t, gpxDoc.Tracks[0].Comment, "example cmt t")
	assertEquals(t, gpxDoc.Tracks[0].Description, "example desc t")
	assertEquals(t, gpxDoc.Tracks[0].Source, "example src t")
	assertEquals(t, gpxDoc.Tracks[0].Number, 1)
	assertEquals(t, gpxDoc.Tracks[0].Type, "")
	// TODO link

	assertEquals(t, len(gpxDoc.Tracks[0].Segments), 2)

	assertEquals(t, len(gpxDoc.Tracks[0].Segments[0].Points), 1)
	assertEquals(t, len(gpxDoc.Tracks[0].Segments[1].Points), 0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Elevation.Value(), 11.1)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Timestamp.Format(TIME_FORMAT), "2013-01-01T12:00:04Z")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].MagneticVariation, "12")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].GeoidHeight, "13")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Name, "example name t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Comment, "example cmt t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Description, "example desc t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Source, "example src t")
	// TODO link
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Symbol, "example sym t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Type, "example type t")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].TypeOfGpsFix, "3d")
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].Satellites.Value(), 100)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].HorizontalDilution.Value(), 101.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].VerticalDilution.Value(), 102.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].PositionalDilution.Value(), 103.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].AgeOfDGpsData.Value(), 104.0)
	assertEquals(t, gpxDoc.Tracks[0].Segments[0].Points[0].DGpsId.Value(), 99)
	// TODO extensions
}

func TestLength2DSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")

	fmt.Println("tracks=", g.Tracks)
	fmt.Println("tracks=", len(g.Tracks))
	fmt.Println("segments=", len(g.Tracks[0].Segments))

	lengthA := g.Tracks[0].Segments[0].Length2D()
	lengthE := 56.77577732775905

	if lengthA != lengthE {
		t.Errorf("Length 2d expected: %f, actual %f", lengthE, lengthA)
	}
}

func TestLength3DSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	lengthA := g.Tracks[0].Segments[0].Length3D()
	lengthE := 61.76815317436073

	if lengthA != lengthE {
		t.Errorf("Length 3d expected: %f, actual %f", lengthE, lengthA)
	}
}

func TestTimePoint(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	timeA := g.Tracks[0].Segments[0].Points[0].Timestamp
	//2012-03-17T12:46:19Z
	timeE := time.Date(2012, 3, 17, 12, 46, 19, 0, time.UTC)

	if timeA != timeE {
		t.Errorf("Time expected: %s, actual: %s", timeE.String(), timeA.String())
	}
}

func TestTimeBoundsSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	timeBoundsA := g.Tracks[0].Segments[0].TimeBounds()

	startTime := time.Date(2012, 3, 17, 12, 46, 19, 0, time.UTC)
	endTime := time.Date(2012, 3, 17, 12, 47, 23, 0, time.UTC)
	timeBoundsE := TimeBounds{
		StartTime: startTime,
		EndTime:   endTime,
	}

	if !timeBoundsE.Equals(timeBoundsA) {
		t.Errorf("TimeBounds expected: %s, actual: %s", timeBoundsE.String(), timeBoundsA.String())
	}
}

func TestBoundsSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")

	boundsA := g.Tracks[0].Segments[0].Bounds()
	boundsE := GpxBounds{
		MaxLat: 52.5117189623, MinLat: 52.5113534275,
		MaxLon: 13.4571944922, MinLon: 13.4567520116,
	}

	if !boundsE.Equals(boundsA) {
		t.Errorf("Bounds expected: %s, actual: %s", boundsE.String(), boundsA.String())
	}
}

func TestBoundsGpx(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")

	boundsA := g.Bounds()
	boundsE := GpxBounds{
		MaxLat: 52.5117189623, MinLat: 52.5113534275,
		MaxLon: 13.4571944922, MinLon: 13.4567520116,
	}

	if !boundsE.Equals(boundsA) {
		t.Errorf("Bounds expected: %s, actual: %s", boundsE.String(), boundsA.String())
	}
}

func TestSpeedSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	speedA := g.Tracks[0].Segments[0].Speed(2)
	speedE := 1.5386074011963367

	if speedE != speedA {
		t.Errorf("Speed expected: %f, actual: %f", speedE, speedA)
	}
}

func TestSegmentDuration(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	durE := 64.0
	durA := g.Tracks[0].Segments[0].Duration()
	if durE != durA {
		t.Errorf("Duration expected: %f, actual: %f", durE, durA)
	}
}

func TestTrackDuration(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	durE := 64.0
	durA := g.Duration()
	if durE != durA {
		t.Errorf("Duration expected: %f, actual: %f", durE, durA)
	}
}

func TestMultiSegmentDuration(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	g.Tracks[0].AppendSegment(g.Tracks[0].Segments[0])
	durE := 64.0 * 2
	durA := g.Duration()
	if durE != durA {
		t.Errorf("Duration expected: %f, actual: %f", durE, durA)
	}
}

func TestMultiTrackDuration(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")

	g.Tracks[0].AppendSegment(g.Tracks[0].Segments[0])
	g.AppendTrack(&g.Tracks[0])
	g.Tracks[0].AppendSegment(g.Tracks[0].Segments[0])

	//xml, _ := g.ToXml(ToXmlParams{Indent: true})
	//fmt.Println(string(xml))

	durE := 320.0
	durA := g.Duration()
	if durE != durA {
		t.Errorf("Duration expected: %f, actual: %f", durE, durA)
	}
}

func TestUphillDownHillSeg(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	updoA := g.Tracks[0].Segments[0].UphillDownhill()
	updoE := UphillDownhill{
		Uphill:   5.863000000000007,
		Downhill: 1.5430000000000064}

	if !updoE.Equals(updoA) {
		t.Errorf("UphillDownhill expected: %+v, actual: %+v", updoE, updoA)
	}
}

func TestMovingData(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	movDataA := g.MovingData()
	movDataE := MovingData{
		MovingTime:      39.0,
		StoppedTime:     25.0,
		MovingDistance:  55.28705571308896,
		StoppedDistance: 6.481097461271765,
		MaxSpeed:        0.0,
	}

	if !movDataE.Equals(movDataA) {
		t.Errorf("Moving data expected: %+v, actual: %+v", movDataE, movDataA)
	}
}

func TestUphillDownhill(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	updoA := g.UphillDownhill()
	updoE := UphillDownhill{
		Uphill:   5.863000000000007,
		Downhill: 1.5430000000000064}

	if !updoE.Equals(updoA) {
		t.Errorf("UphillDownhill expected: %+v, actual: %+v", updoE, updoA)
	}
}

func TestToXml(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	xml, _ := g.ToXml(ToXmlParams{Version: "1.1", Indent: true})
	xmlA := string(xml)
	xmlE := `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" version="1.1" creator="eTrex 10">
	<metadata>
        <author></author>
		<link href="http://www.garmin.com">
			<text>Garmin International</text>
		</link>
		<time>2012-03-17T15:44:18Z</time>
	</metadata>
	<wpt lat="37.085751" lon="-121.17042">
		<ele>195.440933</ele>
		<time>2012-03-21T21:24:43Z</time>
		<name>001</name>
		<sym>Flag, Blue</sym>
	</wpt>
	<wpt lat="37.085751" lon="-121.17042">
		<ele>195.438324</ele>
		<time>2012-03-21T21:24:44Z</time>
		<name>002</name>
		<sym>Flag, Blue</sym>
	</wpt>
	<trk>
		<name>17-MRZ-12 16:44:12</name>
		<trkseg>
			<trkpt lat="52.5113534275" lon="13.4571944922">
				<ele>59.26</ele>
				<time>2012-03-17T12:46:19Z</time>
			</trkpt>
			<trkpt lat="52.5113568641" lon="13.4571697656">
				<ele>65.51</ele>
				<time>2012-03-17T12:46:44Z</time>
			</trkpt>
			<trkpt lat="52.511710329" lon="13.456941694">
				<ele>65.99</ele>
				<time>2012-03-17T12:47:01Z</time>
			</trkpt>
			<trkpt lat="52.5117189623" lon="13.4567520116">
				<ele>63.58</ele>
				<time>2012-03-17T12:47:23Z</time>
			</trkpt>
		</trkseg>
	</trk>
</gpx>`

	assertLinesEquals(t, xmlE, xmlA)
}

func TestNewXml(t *testing.T) {
	gpx := new(GPX)
	gpxTrack := new(GPXTrack)

	gpxSegment := new(GPXTrackSegment)
	gpxSegment.Points = append(gpxSegment.Points, &GPXPoint{Point: Point{Latitude: 2.1234, Longitude: 5.1234, Elevation: *NewNullableFloat64(1234.0)}})
	gpxSegment.Points = append(gpxSegment.Points, &GPXPoint{Point: Point{Latitude: 2.1233, Longitude: 5.1235, Elevation: *NewNullableFloat64(1235.0)}})
	gpxSegment.Points = append(gpxSegment.Points, &GPXPoint{Point: Point{Latitude: 2.1235, Longitude: 5.1236, Elevation: *NewNullableFloat64(1236.0)}})

	gpxTrack.Segments = append(gpxTrack.Segments, gpxSegment)
	gpx.Tracks = append(gpx.Tracks, *gpxTrack)

	xml, _ := gpx.ToXml(ToXmlParams{Version: "1.1", Indent: true})
	actualXml := string(xml)
	// TODO: xsi namespace:
	//expectedXml := `<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.topografix.com/GPX/1/1 http://www.topografix.com/GPX/1/1/gpx.xsd" version="1.1" creator="https://github.com/ptrv/go-gpx">
	expectedXml := `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" version="1.1" creator="https://github.com/ptrv/go-gpx">
	<metadata>
			<author></author>
	</metadata>
	<trk>
		<trkseg>
			<trkpt lat="2.1234" lon="5.1234">
				<ele>1234</ele>
			</trkpt>
			<trkpt lat="2.1233" lon="5.1235">
				<ele>1235</ele>
			</trkpt>
			<trkpt lat="2.1235" lon="5.1236">
				<ele>1236</ele>
			</trkpt>
		</trkseg>
	</trk>
</gpx>`

	assertLinesEquals(t, expectedXml, actualXml)
}

func TestInvalidXML(t *testing.T) {
	xml := "<gpx></gpx"
	gpx, err := ParseString(xml)
	if err == nil {
		t.Error("No error for invalid XML!")
	}
	if gpx != nil {
		t.Error("No gpx should be returned for invalid XMLs")
	}
}

func TestAddElevation(t *testing.T) {
	gpx := new(GPX)
	gpx.AppendTrack(new(GPXTrack))
	gpx.Tracks[0].AppendSegment(new(GPXTrackSegment))
	gpx.Tracks[0].Segments[0].AppendPoint(&GPXPoint{Point: Point{Latitude: 12, Longitude: 13, Elevation: *NewNullableFloat64(100)}})
	gpx.Tracks[0].Segments[0].AppendPoint(&GPXPoint{Point: Point{Latitude: 12, Longitude: 13}})

	gpx.AddElevation(10)
	assertEquals(t, gpx.Tracks[0].Segments[0].Points[0].Elevation.Value(), 110.0)
	assertEquals(t, gpx.Tracks[0].Segments[0].Points[1].Elevation.Value(), 10.0) // TODO: this should be nil!

	gpx.AddElevation(-20)
	assertEquals(t, gpx.Tracks[0].Segments[0].Points[0].Elevation.Value(), 90.0)
	assertEquals(t, gpx.Tracks[0].Segments[0].Points[1].Elevation.Value(), -10.0) // TODO: this should be nil!
}

func TestRemoveElevation(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")

	// Remove elevations don't work on waypoints and routes, so just remove them for this test (TODO)
	g.Waypoints = make([]GPXPoint, 0)
	g.Routes = make([]GPXRoute, 0)

	{
		xml, _ := g.ToXml(ToXmlParams{Indent: true})
		if !strings.Contains(string(xml), "<ele") {
			t.Error("No elevations for the test")
		}
	}

	g.RemoveElevation()

	{
		xml, _ := g.ToXml(ToXmlParams{Indent: true})

		//fmt.Println(string(xml))

		if strings.Contains(string(xml), "<ele") {
			t.Error("Elevation still there!")
		}
	}
}

func TestExecuteOnAllPoints(t *testing.T) {
	g, _ := ParseFile("../test_files/file.gpx")
	g.ExecuteOnAllPoints(func(*GPXPoint) {
	})
}

/* TODO
func TestEmptyElevation(t *testing.T) {
	gpx := new(GPX)
	gpx.AppendTrack(new(GPXTrack))
	gpx.Tracks[0].AppendSegment(new(GPXTrackSegment))
	gpx.Tracks[0].Segments[0].AppendPoint(&GPXPoint{Point: Point{Latitude: 12, Longitude: 13, Elevation: 100}})
	gpx.Tracks[0].Segments[0].AppendPoint(&GPXPoint{Point: Point{Latitude: 13, Longitude: 14, Elevation: 0}})
	gpx.Tracks[0].Segments[0].AppendPoint(&GPXPoint{Point: Point{Latitude: 14, Longitude: 15}})

	xmlBytes, _ := gpx.ToXml(ToXmlParams{Indent: false})
	xml := string(xmlBytes)

	if !strings.Contains(xml, `<trkpt lat="12" lon="13"><ele>100</ele></trkpt>`) {
		t.Error("Invalid elevation 100 serialization:" + xml)
	}
	if !strings.Contains(xml, `<trkpt lat="13" lon="14"><ele>0</ele></trkpt>`) {
		t.Error("Invalid elevation 0 serialization:" + xml)
	}
	if !strings.Contains(xml, `<trkpt lat="14" lon="15"></trkpt>`) {
		t.Error("Invalid empty elevation serialization:" + xml)
	}
}
*/

// TODO:
// RemoveTime

func TestTrackWithoutTimes(t *testing.T) {
	g, _ := ParseFile("../test_files/cerknicko-without-times.gpx")
	if g.HasTimes() {
		t.Error("Track should not have times")
	}
}

/*
   def test_has_times_false(self):
       gpx = self.parse('cerknicko-without-times.gpx')
       self.assertFalse(gpx.has_times())

   def test_has_times(self):
       gpx = self.parse('korita-zbevnica.gpx')
       self.assertTrue(len(gpx.tracks) == 4)
       # Empty -- True
       self.assertTrue(gpx.tracks[0].has_times())
       # Not times ...
       self.assertTrue(not gpx.tracks[1].has_times())

       # Times OK
       self.assertTrue(gpx.tracks[2].has_times())
       self.assertTrue(gpx.tracks[3].has_times())
*/

//func TestHasTimes(t *testing.T) {}

func testReduceTrackByMaxPoints(t *testing.T, maxReducedPointsNo int) {
	for _, gpxFile := range loadTestGPXs() {
		g, _ := ParseFile(gpxFile)
		pointsOriginal := g.GetTrackPointsNo()

		//fmt.Printf("reducing %s to %d points", gpxFile, maxReducedPointsNo)
		g.ReduceTrackPoints(maxReducedPointsNo, 0)

		pointsReduced := g.GetTrackPointsNo()

		if pointsReduced > pointsOriginal {
			//fmt.Printf("Points before %d, now %d\n", pointsOriginal, pointsReduced)
			t.Error("Reduced track has no reduced number of points")
		}
	}
}

func testReduceTrackByMaxPointsAndMinDistance(t *testing.T, maxReducedPointsNo int, minDistance float64) {
	for _, gpxFile := range loadTestGPXs() {
		g, _ := ParseFile(gpxFile)
		pointsOriginal := g.GetTrackPointsNo()

		//fmt.Printf("reducing %s to %d points and min distance %f\n", gpxFile, maxReducedPointsNo, minDistance)
		g.ReduceTrackPoints(maxReducedPointsNo, minDistance)

		minDistanceOriginal := getMinDistanceBetweenTrackPoints(*g)

		pointsReduced := g.GetTrackPointsNo()
		if pointsReduced > pointsOriginal {
			//fmt.Printf("Points before %d, now %d\n", pointsOriginal, pointsReduced)
			t.Error("Reduced track has no reduced number of points")
		}

		reducedMinDistance := getMinDistanceBetweenTrackPoints(*g)
		//fmt.Printf("fileName=%s after reducing pointsNo=%d, minDistance=%f\n", gpxFile, g.GetTrackPointsNo(), reducedMinDistance)

		if minDistanceOriginal > 0.0 {
			if reducedMinDistance < minDistance {
				t.Error(fmt.Sprintf("reducedMinDistance=%f, but minDistance should be=%f", reducedMinDistance, minDistance))
			}
		}
	}
}

func TestReduceTrackByMaxPointsAndMinDistance(t *testing.T) {
	testReduceTrackByMaxPointsAndMinDistance(t, 100000, 10.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 100000, 50.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 100000, 200.0)

	testReduceTrackByMaxPointsAndMinDistance(t, 10, 0.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 100, 0.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 1000, 0.0)

	testReduceTrackByMaxPointsAndMinDistance(t, 10, 10.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 100, 10.0)
	testReduceTrackByMaxPointsAndMinDistance(t, 1000, 20.0)
}

func simplifyAndCheck(t *testing.T, gpxFile string, maxDistance float64) float64 {
	g, _ := ParseFile(gpxFile)
	length2DOriginal := g.Length2D()

	g.SimplifyTracks(maxDistance)

	length2DAfterSimplified := g.Length2D()
	if length2DOriginal < length2DAfterSimplified {
		t.Error(fmt.Sprintf("Original length cannot be smaller than simplified, original=%f, simplified=%f", length2DOriginal, length2DAfterSimplified))
	}

	return length2DAfterSimplified
}

func TestSimplifyForSingleSegmentAndVeryByMaxDistance(t *testing.T) {
	g, _ := ParseFile("../test_files/Mojstrovka.gpx")

	assertTrue(t, "Single track, single segment track for this test", len(g.Tracks) == 1 && len(g.Tracks[0].Segments) == 1)
	assertTrue(t, "More than 2 points needed", g.GetTrackPointsNo() > 2)

	g.SimplifyTracks(1000000000.0)

	assertTrue(t, fmt.Sprintf("maxDistance very big => only first and last points should be left, found:%d", g.GetTrackPointsNo()), g.GetTrackPointsNo() == 2)

	start := g.Tracks[0].Segments[0].Points[0]
	end := g.Tracks[0].Segments[0].Points[len(g.Tracks[0].Segments[0].Points)-1]
	distanceBetweenFirstAndLast := start.Distance2D(&end.Point)
	assertTrue(t, fmt.Sprintf("maxDistance very big => only first and last points should be left %f!=%f", g.Length2D(), distanceBetweenFirstAndLast), cca(g.Length2D(), distanceBetweenFirstAndLast))
}

func TestSimplify(t *testing.T) {
	for _, gpxFile := range loadTestGPXs() {
		g, _ := ParseFile(gpxFile)

		length2dAfterMaxDistance10000000000 := simplifyAndCheck(t, gpxFile, 10000000000.0)
		length2dAfterMaxDistance50 := simplifyAndCheck(t, gpxFile, 50.0)
		length2dAfterMaxDistance10 := simplifyAndCheck(t, gpxFile, 10.0)
		length2dAfterMaxDistance5 := simplifyAndCheck(t, gpxFile, 5.0)
		length2dAfterMaxDistance0000001 := simplifyAndCheck(t, gpxFile, 0.000001)

		/*
		   fmt.Println()
		   fmt.Println("length2dAfterMaxDistance10000000000=", length2dAfterMaxDistance10000000000)
		   fmt.Println("length2dAfterMaxDistance50=", length2dAfterMaxDistance50)
		   fmt.Println("length2dAfterMaxDistance10=", length2dAfterMaxDistance10)
		   fmt.Println("length2dAfterMaxDistance5=", length2dAfterMaxDistance5)
		   fmt.Println("length2dAfterMaxDistance0000001=", length2dAfterMaxDistance0000001)
		*/

		assertTrue(t, "If maxDistance very small then the simplified length should be almost the same thant the original", length2dAfterMaxDistance0000001 == g.Length2D())

		assertTrue(t, "Bigger maxDistance => smaller simplified track", length2dAfterMaxDistance10000000000 <= length2dAfterMaxDistance50)
		assertTrue(t, "Bigger maxDistance => smaller simplified track", length2dAfterMaxDistance50 <= length2dAfterMaxDistance10)
		assertTrue(t, "Bigger maxDistance => smaller simplified track", length2dAfterMaxDistance10 <= length2dAfterMaxDistance5)
		assertTrue(t, "Bigger maxDistance => smaller simplified track", length2dAfterMaxDistance5 <= length2dAfterMaxDistance0000001)
	}
}
