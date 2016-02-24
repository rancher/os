package selinux

// SetFileContext is a stub for SELinux support on ARM
func SetFileContext(path string, context string) (int, error) {
	return 0, nil
}
