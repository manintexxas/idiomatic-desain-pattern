package normalize

func isDangerousExt(s string) bool {
	switch s {
	case "php", "phtml", "pht", "php3", "php4", "php5", "php7", "php8":
		return true

	case "asp", "aspx", "ashx", "asmx", "asa", "asax":
		return true

	case "jsp", "jspx":
		return true

	case "shtml":
		return true

	case "exe", "com", "scr", "msi", "bat", "cmd", "ps1", "vbs", "vbe", "wsf", "wsh","hta":
		return true
	}

	return false
}
