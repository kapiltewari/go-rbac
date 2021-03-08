package utils

//CheckExtension of an image
func CheckExtension(extension string) bool {
	allowedExtensions := []string{".jpg", ".jpeg", ".png"}

	for _, ext := range allowedExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}
