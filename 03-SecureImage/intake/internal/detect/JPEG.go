package detect

import (
	in "03-SecureImage/intake"
)

func signatureJPEG(b []byte) (bool, error) {
	if len(b) < len(SOI) {
		return false, in.ErrSignatureTooShort
	}
	
	for i := range SOI {
		if b[i] != SOI[i] {
			return false, nil
		}
	}

	return true, nil
}