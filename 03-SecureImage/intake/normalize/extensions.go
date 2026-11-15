package normalize

var (
	dangerousIntermediateExtensions = [...]string{
		"php", "phtml", "pht",
		"php3", "php4", "php5", "php7", "php8",

		"asp", "aspx", "ashx", "asmx",
		"asa", "asax",

		"jsp", "jspx",

		"shtml",

		"exe", "com", "scr", "msi",
		"bat", "cmd", "ps1",
		"vbs", "vbe", "wsf", "wsh", "hta",
	}

	ext = [4]string{"jpg", "jpeg", "png", "webp"}
)
