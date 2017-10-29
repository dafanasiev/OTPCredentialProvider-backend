package shared

import (
	"strings"
)


type PathResolver interface{
	PathToAbs(path string) string
}

type pathResoler struct{
	appRootDir string
	dataDir    string
	configDir  string
}

func NewPathResolver(appRootDir string, dataDir string, configDir string) PathResolver {
	return &pathResoler{
		appRootDir: appRootDir,
		dataDir:    dataDir,
		configDir:  configDir,
	}
}

func (p *pathResoler)PathToAbs(pathStr string) string {

	pathStr = strings.Replace(pathStr, "${dir.bin}", p.appRootDir, -1)
	pathStr = strings.Replace(pathStr, "${dir.data}", p.dataDir, -1)
	pathStr = strings.Replace(pathStr, "${dir.config}", p.configDir, -1)

	return pathStr
}