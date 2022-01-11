package config

func NewConfig() *Config {
	c := new(Config)
	{
		c.Application.Quiet = false
		c.Application.WorkingDir = "."
		c.Application.CacheDir = ""
	}
	{
		c.Minecraft.Quiet = false
		c.Minecraft.NoGUI = false
		c.Minecraft.NoEULA = false
	}
	{
		c.Java.ExecutableOverride = ""
		c.Java.Memory.Amount = 6
		c.Java.Memory.Gigabytes = true
		c.Java.Flags.ExtraFlags = nil
		c.Java.Flags.OverrideDefault = false
	}
	{
		c.Git.Disable = false
		c.Git.DisableGithubIntegration = false
		c.Git.Overrides.Enable = false
		c.Git.Overrides.CustomPreCommands = nil
		c.Git.Overrides.CustomPostCommands = nil
	}
	return c
}
