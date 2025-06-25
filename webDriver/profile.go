package webdriver

import (
	"PhoneNumberCheck/utils"
	"os"
	"path/filepath"
)

// func base64EncodeProfile() (string, error) {
// 	buf := new(bytes.Buffer)
// 	zipWriter := zip.NewWriter(buf)
//
// 	err := filepath.Walk(profileDir, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
//
// 		if info.IsDir() {
// 			return nil
// 		}
//
// 		relPath, err := filepath.Rel(profileDir, path)
// 		if err != nil {
// 			return err
// 		}
//
// 		zipFile, err := zipWriter.Create(relPath)
// 		if err != nil {
// 			return err
// 		}
//
// 		srcFile, err := os.Open(path)
// 		if err != nil {
// 			return err
// 		}
// 		defer srcFile.Close()
//
// 		_, err = io.Copy(zipFile, srcFile)
// 		return err
// 	})
//
// 	if err != nil {
// 		return "", err
// 	}
//
// 	err = zipWriter.Close()
// 	if err != nil {
// 		return "", err
// 	}
//
// 	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
// 	return encoded, nil
// }

func createProfile(providerName WebScrapingProvider) (providerProfilePath string, err error) {

	baseProfileDir, err := filepath.Abs(".firefoxProfiles")
	if err != nil {
		return "", err
	}

	providerProfilePath = filepath.Join(baseProfileDir, string(providerName))
	if exists := utils.CheckIfFileExists(providerProfilePath); exists {
		return providerProfilePath, nil
	}

	err = os.MkdirAll(providerProfilePath, 0644)
	if err != nil {
		return providerProfilePath, err
	}

	prefs := `
		user_pref("browser.shell.checkDefaultBrowser", false);
		user_pref("browser.startup.homepage", "about:blank");
		user_pref("startup.homepage_welcome_url", "about:blank");
		user_pref("startup.homepage_welcome_url.additional", "about:blank");
		user_pref("browser.startup.firstrunSkipsHomepage", true);
		user_pref("browser.startup.page", 0);
		user_pref("browser.newtabpage.enabled", false);
		user_pref("datareporting.healthreport.uploadEnabled", false);
		user_pref("datareporting.policy.dataSubmissionEnabled", false);
		user_pref("toolkit.telemetry.enabled", false);
		user_pref("toolkit.telemetry.unified", false);
		user_pref("toolkit.telemetry.archive.enabled", false);
		user_pref("toolkit.telemetry.prompted", 2);
		user_pref("toolkit.telemetry.rejected", true);
		user_pref("browser.cache.disk.enable", true);
		user_pref("browser.cache.memory.enable", true);
		user_pref("browser.cache.disk.capacity", 1048576); // 1GB
		user_pref("permissions.default.image", 2); // Disable image loading
		user_pref("gfx.downloadable_fonts.enabled", false);
		user_pref("privacy.clearOnShutdown.cookies", false);
		user_pref("signon.rememberSignons", false);
		user_pref("network.cookie.lifetimePolicy", 0); // Keep cookies until they expire
`
	prefsPath := filepath.Join(baseProfileDir, "prefs.js")
	err = os.WriteFile(prefsPath, []byte(prefs), 0644)
	if err != nil {
		return providerProfilePath, err
	}

	return
}
