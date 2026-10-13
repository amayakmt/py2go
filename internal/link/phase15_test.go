package link

import "testing"

func TestLinkJoinsTheFormatModules(t *testing.T) {
	mod, err := project(t, map[string]string{"main.py": "import binascii\nimport base64\nimport hashlib\nimport hmac\nimport uuid\nimport csv\nimport configparser\nimport html\nimport html.parser\nimport zlib\nimport gzip\nimport quopri\nimport pickle\nimport copyreg\n\nprint(binascii.hexlify(b'a'), base64.b64encode(b'a'), hashlib.blake2b(), hmac.new, uuid.uuid4(), csv.reader, configparser.ConfigParser(), html.escape('<'), html.parser.HTMLParser(), zlib.crc32(b'a'), gzip.compress(b'a'), quopri.encodestring(b'a'), pickle.dumps(1), copyreg.dispatch_table)\n"})
	if err != nil {
		t.Fatal(err)
	}
	got := names(mod)
	for _, want := range []string{"binascii_pm_hexlify", "base64_pm_b64encode", "hashlib_pm_blake2b", "hmac_pm_new", "uuid_pm_uuid4", "csv_pm_reader", "configparser_pm_ConfigParser", "html_pm_escape", "html__parser_pm_HTMLParser", "zlib_pm_crc32", "gzip_pm_compress", "quopri_pm_encodestring", "pickle_pm_dumps", "copyreg_pm_dispatch_table"} {
		if !got[want] {
			t.Errorf("%s is not in the joined module", want)
		}
	}
}
