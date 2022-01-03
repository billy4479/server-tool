package main

func gitPreFn(s *Server) (err error) {
	_, err = runCmdPretty(false, true, s.BaseDir, "git", "pull")
	return err
}

func gitPostFn(s *Server) (err error) {
	_, err = runCmdPretty(false, true, s.BaseDir, "git", "add", "-A")
	if err != nil {
		return err
	}

	_, err = runCmdPretty(false, true, s.BaseDir, "git", "commit", "--allow-empty-message", "-m", "")
	if err != nil {
		return err
	}

	_, err = runCmdPretty(false, true, s.BaseDir, "git", "push")
	return err
}
