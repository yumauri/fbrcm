package theme

import "os"

func openThemeSource(path string) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, &os.PathError{Op: "read", Path: path, Err: os.ErrInvalid}
	}
	return file, nil
}
