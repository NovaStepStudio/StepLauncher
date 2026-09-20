package helpers

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"StepLauncher/internal/Core/Utils"
)

type VarConfig struct {
	Username, UUID, AccessToken, XUID, ClientID string
	GameDir, AssetsDir, LibrariesDir            string
	LauncherName, LauncherVersion               string
	DemoUser, CustomResolution                  bool
	ResWidth, ResHeight                         int
	UserType                                    string
	GCArgs                                      []string
	HWAccelFlags                                []string
	LauncherProperties                          string
}

func BuildVarsMap(cfg VarConfig, verID, verType, assetsIndexID, classpath, nativesDir string) map[string]string {
	if verType == "" {
		verType = "release"
	}
	if assetsIndexID == "" {
		assetsIndexID = "index"
	}

	sep := ";"
	if runtime.GOOS != "windows" {
		sep = ":"
	}

	m := map[string]string{
		"auth_player_name":    cfg.Username,
		"auth_access_token":   cfg.AccessToken,
		"auth_uuid":           cfg.UUID,
		"auth_xuid":           cfg.XUID,
		"auth_session":        cfg.AccessToken,
		"clientid":            cfg.ClientID,
		"game_directory":      cfg.GameDir,
		"game_assets":         cfg.AssetsDir,
		"assets_root":         cfg.AssetsDir,
		"assets_index_name":   assetsIndexID,
		"version_name":        verID,
		"version_type":        verType,
		"natives_directory":   nativesDir,
		"launcher_name":       cfg.LauncherName,
		"launcher_version":    cfg.LauncherVersion,
		"classpath":           classpath,
		"classpath_separator": sep,
		"library_directory":   cfg.LibrariesDir,
		"user_type":           orDefault(cfg.UserType, "mojang"),
		"user_properties":     "{}",
		"profile_name":        verID,
		"resolution_width":    strconv.Itoa(cfg.ResWidth),
		"resolution_height":   strconv.Itoa(cfg.ResHeight),
		"launcher_properties": cfg.LauncherProperties,
	}
	return m
}

func BuildFeaturesMap(demoUser, customResolution bool) map[string]bool {
	return map[string]bool{
		"is_demo_user":          demoUser,
		"has_custom_resolution": customResolution,
	}
}

type JVMArgResult struct {
	Args  []string
	Extra []string
}

func BuildJVMArgs(rawArgs []interface{}, vars map[string]string) []string {
	var result []string
	for _, arg := range rawArgs {
		switch v := arg.(type) {
		case string:
			r := SubstituteVars(v, vars)
			if strings.TrimSpace(r) != "" {
				result = append(result, r)
			}
		case map[string]interface{}:
			rulesRaw, hasRules := v["rules"]
			if !hasRules {
				continue
			}
			if !EvaluateRules(rulesRaw, nil) {
				continue
			}
			valueRaw, hasValue := v["value"]
			if !hasValue {
				continue
			}
			result = append(result, expandValue(valueRaw, vars)...)
		}
	}
	return result
}

func BuildGameArgs(rawArgs []interface{}, minecraftArgs string, vars map[string]string, features map[string]bool, fullscreen bool) []string {
	var result []string

	if len(rawArgs) > 0 {
		for _, arg := range rawArgs {
			switch v := arg.(type) {
			case string:
				r := SubstituteVars(v, vars)
				if strings.TrimSpace(r) != "" {
					result = append(result, r)
				}
			case map[string]interface{}:
				rulesRaw, hasRules := v["rules"]
				if !hasRules {
					continue
				}
				if !EvaluateRules(rulesRaw, features) {
					continue
				}
				valueRaw, hasValue := v["value"]
				if !hasValue {
					continue
				}
				result = append(result, expandValue(valueRaw, vars)...)
			}
		}
	} else if minecraftArgs != "" {
		parsed := SubstituteVars(minecraftArgs, vars)
		result = strings.Fields(parsed)
	}

	if fullscreen {
		result = append(result, "--fullscreen")
	}

	return result
}

func expandValue(valueRaw interface{}, vars map[string]string) []string {
	var result []string
	switch val := valueRaw.(type) {
	case string:
		r := SubstituteVars(val, vars)
		if strings.TrimSpace(r) != "" {
			result = append(result, r)
		}
	case []interface{}:
		for _, item := range val {
			if s, ok := item.(string); ok {
				r := SubstituteVars(s, vars)
				if strings.TrimSpace(r) != "" {
					result = append(result, r)
				}
			}
		}
	}
	return result
}

func SubstituteVars(template string, vars map[string]string) string {
	if template == "" || !strings.Contains(template, "${") {
		return template
	}
	current := template
	for pass := 0; pass < 8; pass++ {
		if !strings.Contains(current, "${") {
			break
		}
		var b strings.Builder
		b.Grow(len(current))
		replaced := false
		rest := current
		for {
			start := strings.Index(rest, "${")
			if start < 0 {
				b.WriteString(rest)
				break
			}
			closeIdx := strings.IndexByte(rest[start+2:], '}')
			if closeIdx < 0 {
				b.WriteString(rest)
				break
			}
			key := rest[start+2 : start+2+closeIdx]
			val, ok := vars[key]
			b.WriteString(rest[:start])
			if ok {
				b.WriteString(val)
				replaced = true
			} else {
				b.WriteString(rest[start : start+2+closeIdx+1])
			}
			rest = rest[start+2+closeIdx+1:]
		}
		current = b.String()
		if !replaced {
			break
		}
	}
	return current
}

func EvaluateRules(rulesRaw interface{}, features map[string]bool) bool {
	ruleList, ok := rulesRaw.([]interface{})
	if !ok || len(ruleList) == 0 {
		return true
	}

	hasAllow := false
	for _, r := range ruleList {
		rule, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		if rule["action"] == "allow" {
			hasAllow = true
			break
		}
	}
	if !hasAllow {
		for _, r := range ruleList {
			rule, ok := r.(map[string]interface{})
			if !ok {
				continue
			}
			if rule["action"] == "disallow" && evalRuleConditions(rule, features) {
				return false
			}
		}
		return true
	}
	allow := false
	for _, r := range ruleList {
		rule, ok := r.(map[string]interface{})
		if !ok {
			continue
		}
		action, _ := rule["action"].(string)
		switch action {
		case "allow":
			if evalRuleConditions(rule, features) {
				allow = true
			}
		case "disallow":
			if evalRuleConditions(rule, features) {
				return false
			}
		}
	}
	return allow
}

func evalRuleConditions(rule map[string]interface{}, features map[string]bool) bool {
	if osRaw, ok := rule["os"]; ok {
		osMap, ok := osRaw.(map[string]interface{})
		if !ok || !matchOS(osMap) {
			return false
		}
	}
	if featRaw, ok := rule["features"]; ok && features != nil {
		featMap, ok := featRaw.(map[string]interface{})
		if !ok || !matchFeatures(featMap, features) {
			return false
		}
	}
	return true
}

func matchOS(osMap map[string]interface{}) bool {
	if name, ok := osMap["name"].(string); ok && name != "" {
		if name != utils.OsName() {
			return false
		}
	}
	if arch, ok := osMap["arch"].(string); ok && arch != "" {
		if arch != utils.OsArch() {
			return false
		}
	}
	if verRaw, ok := osMap["version"]; ok {
		switch v := verRaw.(type) {
		case string:
			if !matchOSVersionString(v) {
				return false
			}
		case map[string]interface{}:
			if !matchOSVersionRange(v) {
				return false
			}
		}
	}
	return true
}

func matchOSVersionString(version string) bool {
	return strings.Contains(runtime.GOOS, version) || strings.Contains(os.Getenv("OS"), version)
}

func matchOSVersionRange(vr map[string]interface{}) bool {
	if runtime.GOOS != "windows" {
		return true
	}
	minVer, hasMin := vr["min"].(string)
	maxVer, hasMax := vr["max"].(string)
	if !hasMin && !hasMax {
		return true
	}
	winVer := windowsVersion()
	if hasMin && compareVersions(winVer, minVer) < 0 {
		return false
	}
	if hasMax && compareVersions(winVer, maxVer) > 0 {
		return false
	}
	return true
}

func matchFeatures(featMap map[string]interface{}, features map[string]bool) bool {
	for key, expectedRaw := range featMap {
		expected, ok := expectedRaw.(bool)
		if !ok {
			return false
		}
		actual, exists := features[key]
		if !exists || actual != expected {
			return false
		}
	}
	return true
}

func windowsVersion() string {
	return os.Getenv("OS")
}

func compareVersions(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < len(ap) && i < len(bp); i++ {
		ai, _ := strconv.Atoi(ap[i])
		bi, _ := strconv.Atoi(bp[i])
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	if len(ap) < len(bp) {
		return -1
	}
	if len(ap) > len(bp) {
		return 1
	}
	return 0
}

func FormatArgsForLog(args []string) string {
	if len(args) > 25 {
		shown := make([]string, 25)
		copy(shown, args[:25])
		return fmt.Sprintf("%s ... [%d total]", strings.Join(shown, " "), len(args))
	}
	return strings.Join(args, " ")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
