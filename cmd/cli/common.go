package cli

import (
	"os"
	"strings"
)

func getDescriptions(descriptionArg []string, _ bool) map[string]string {
	var description, banner string
	if strings.Contains(strings.Join(os.Args[0:], ""), "-h") {
		description = descriptionArg[0]
	} else {
		description = descriptionArg[1]
	}
	//if !hideBanner {
	banner = `   ______      __ __      __              ___________
  / ____/___  / //_/_  __/ /_  ___  _  __/ ____/ ___/
 / / __/ __ \/ ,< / / / / __ \/ _ \| |/_/ /_   \__ \ 
/ /_/ / /_/ / /| / /_/ / /_/ /  __/>  </ __/  ___/ / 
\____/\____/_/ |_\__,_/_.___/\___/_/|_/_/    /____/
`
	//} else {
	//banner = ""
	//}
	return map[string]string{"banner": banner, "description": description}
}
