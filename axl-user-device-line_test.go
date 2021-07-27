package main

import "testing"

func TestNewUserDeviceLineList(t *testing.T) {
	t.Parallel()
	for i, table := range deviceLineTables {
		_, err := NewUserDeviceLineList(table.t)
		if len(table.version) > 0 {
			if err != nil {
				t.Errorf("get data error %s", err)
			}
		} else {
			if err == nil {
				t.Errorf("not get error as expect on test line %d", i)
			}
		}
	}
}

const deviceLineResponse = `<return>
                <row>
                    <user_pkid>00061774-ff1c-4ecf-9577-cfc930d0f164</user_pkid>
                    <device_pkid>0f8cb698-1d51-0102-9ca3-b743746040f1</device_pkid>
                    <line_pkid>58ed24e1-366b-5549-9bd3-d374b79a3c8e</line_pkid>
                    <firstname>Agent59</firstname>
                    <middlename/>
                    <lastname>Group5</lastname>
                    <userid>agent59</userid>
                    <department>Team Group 5</department>
                    <status>1</status>
                    <islocaluser>t</islocaluser>
                    <uccx>t</uccx>
                    <directoryuri>agent59@callrecordlab.com</directoryuri>
                    <mailid/>
                    <devicename>agent02</devicename>
                    <devicedescrition>AAgent 02 - 2102</devicedescrition>
                    <dnorpattern>2102</dnorpattern>
                    <alertingnameascii>(019)2102</alertingnameascii>
                    <cluster_name>UCS-11</cluster_name>
                    <line_description>(019)2102</line_description>
                </row>
                <row>
                    <user_pkid>0666075e-63d9-4e8e-b3f7-af19cd2847f6</user_pkid>
                    <device_pkid>0f8cb698-1d51-0102-9ca3-b743746040f1</device_pkid>
                    <line_pkid>58ed24e1-366b-5549-9bd3-d374b79a3c8e</line_pkid>
                    <firstname>Agent31</firstname>
                    <middlename/>
                    <lastname>Group3</lastname>
                    <userid>agent31</userid>
                    <department>Team Group 3</department>
                    <status>1</status>
                    <islocaluser>t</islocaluser>
                    <uccx>t</uccx>
                    <directoryuri>agent31@callrecordlab.com</directoryuri>
                    <mailid/>
                    <devicename>agent02</devicename>
                    <devicedescrition>AAgent 02 - 2102</devicedescrition>
                    <dnorpattern>2102</dnorpattern>
                    <alertingnameascii>(019)2102</alertingnameascii>
                    <cluster_name>UCS-11</cluster_name>
                    <line_description>(019)2102</line_description>
                </row>
                <row>
                    <user_pkid>0861376b-bc6e-4a83-8527-f42b15c29d98</user_pkid>
                    <device_pkid>0f8cb698-1d51-0102-9ca3-b743746040f1</device_pkid>
                    <line_pkid>58ed24e1-366b-5549-9bd3-d374b79a3c8e</line_pkid>
                    <firstname>Agent51</firstname>
                    <middlename/>
                    <lastname>Group5</lastname>
                    <userid>agent51</userid>
                    <department>Team Group 5</department>
                    <status>1</status>
                    <islocaluser>t</islocaluser>
                    <uccx>t</uccx>
                    <directoryuri>agent51@callrecordlab.com</directoryuri>
                    <mailid/>
                    <devicename>agent02</devicename>
                    <devicedescrition>AAgent 02 - 2102</devicedescrition>
                    <dnorpattern>2102</dnorpattern>
                    <alertingnameascii>(019)2102</alertingnameascii>
                    <cluster_name>UCS-11</cluster_name>
                    <line_description>(019)2102</line_description>
                </row>
                <row>
                    <user_pkid>089de0d7-4e20-4cd0-84db-b8e2e0e8d7d0</user_pkid>
                    <device_pkid>0f8cb698-1d51-0102-9ca3-b743746040f1</device_pkid>
                    <line_pkid>58ed24e1-366b-5549-9bd3-d374b79a3c8e</line_pkid>
                    <firstname>Agent45</firstname>
                    <middlename/>
                    <lastname>Group4</lastname>
                    <userid>agent45</userid>
                    <department>Team Group 4</department>
                    <status>1</status>
                    <islocaluser>t</islocaluser>
                    <uccx>t</uccx>
                    <directoryuri>agent45@callrecordlab.com</directoryuri>
                    <mailid/>
                    <devicename>agent02</devicename>
                    <devicedescrition>AAgent 02 - 2102</devicedescrition>
                    <dnorpattern>2102</dnorpattern>
                    <alertingnameascii>(019)2102</alertingnameascii>
                    <cluster_name>UCS-11</cluster_name>
                    <line_description>(019)2102</line_description>
                </row>
                <row>
                    <user_pkid>0a99bbcb-7799-79d3-0f95-b14956e266b3</user_pkid>
                    <device_pkid>0d01cd93-9e5e-fce4-ab98-5c7ec160c501</device_pkid>
                    <line_pkid>b8baf7f3-84d5-2c11-cb79-51da82d42dd1</line_pkid>
                    <firstname>Rastislav</firstname>
                    <middlename/>
                    <lastname>Skultety</lastname>
                    <userid>rastislav.skultety</userid>
                    <department/>
                    <status>1</status>
                    <islocaluser>t</islocaluser>
                    <uccx>f</uccx>
                    <directoryuri>rastislav.skultety@callrecordlab.com</directoryuri>
                    <mailid/>
                    <devicename>juraj.onuska</devicename>
                    <devicedescrition>Juraj Onuska 0.19.2065</devicedescrition>
                    <dnorpattern>2065</dnorpattern>
                    <alertingnameascii>Juraj Onuska</alertingnameascii>
                    <cluster_name>UCS-11</cluster_name>
                    <line_description>Juraj Onuska</line_description>
                </row>
            </return>`

var deviceLineTables = []struct {
	t         string
	version   string
	dbVersion string
}{
	{deviceLineResponse, "11.5.1.14900(11)", "11.0"},
	{failResponse, "", ""},
}
