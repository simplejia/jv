// A partner for svn command.
// Created by simplejia [11/2016]
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	Checkout  bool
	Switch    bool
	Merge     bool
	Catch     bool
	Branch    bool
	DelBranch bool
)

func fatal(msg string) {
	fmt.Printf("program exit, msg: \n%s\n", msg)
	os.Exit(0)
}

func run(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fatal(err.Error())
	}
}

func pipe(name string, arg ...string) string {
	output, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		fatal(err.Error())
	}
	return string(output)
}

func input(tips string) string {
	fmt.Print(tips)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fatal("abort")
	}
	return strings.TrimSpace(line)
}

func username() string {
	name := strings.TrimSpace(os.Getenv("JV_USER"))
	if name != "" {
		return name
	}
	cur, err := user.Current()
	if err != nil {
		fatal(err.Error())
	}
	return cur.Username
}

func svns() []string {
	str := strings.TrimSpace(os.Getenv("JV_PATHS"))
	if str == "" {
		fatal("must set JV_PATHS env")
	}
	paths := strings.Split(str, ",")
	for n, path := range paths {
		paths[n] = strings.TrimSuffix(path, "/")
	}
	return paths
}

func chooseTrunk() string {
	paths := svns()
	for n, path := range paths {
		fmt.Printf("%d.\t%s\n", n+1, path)
	}
	n, err := strconv.Atoi(input("\nchoose trunk: "))
	if err != nil {
		fatal(err.Error())
	}
	if n < 1 || n > len(paths) {
		fatal("wrong choice")
	}

	return paths[n-1] + "/trunk"
}

func chooseBranch(home string, includeTrunk bool) string {
	branches := []string{}
	output := pipe("svn", "ls", home)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSuffix(strings.TrimSpace(line), "/")
		if line == "" {
			continue
		}
		branches = append(branches, line)
	}
	sort.Strings(branches)

	if includeTrunk {
		branches = append([]string{"/trunk"}, branches...)
	}

	for n, branch := range branches {
		baseMsg := ""
		if branch != "/trunk" {
			baseMsg = getBaseMsg(fmt.Sprintf("%s/%s", home, branch))
		} else {
			baseMsg = "(master)"
		}
		fmt.Printf("%d.\t%s\t%s\n", n+1, branch, baseMsg)
	}

	n, err := strconv.Atoi(input("\nchoose branch: "))
	if err != nil {
		fatal(err.Error())
	}
	if n < 1 || n > len(branches) {
		fatal("wrong choice")
	}

	return branches[n-1]
}

func isModified() bool {
	if strings.TrimSpace(pipe("svn", "st", "-q")) == "" {
		return false
	} else {
		return true
	}
}

func branchExist(url string) bool {
	_, err := exec.Command("svn", "ls", url).CombinedOutput()
	if err != nil {
		return false
	}
	return true
}

func getBaseMsg(url string) string {
	lines := strings.Split(pipe("svn", "log", "--stop-on-copy", url), "\n")
	if len(lines) >= 6 {
		return strings.TrimSpace(lines[len(lines)-3])
	}
	return ""
}

func getLatestRev(url string) string {
	lines := strings.Split(pipe("svn", "log", "-l1", url), "\n")
	if len(lines) != 6 {
		fatal("unexpected error")
	}
	line := lines[1]
	if pos := strings.Index(line, " "); pos != -1 {
		return line[1:pos]
	}
	fatal("unexpected here")
	return ""
}

func getInfos() (trunk, home, root, url string) {
	svninfos := map[string]string{}
	for _, line := range strings.Split(pipe("svn", "info"), "\n") {
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		k, v := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
		svninfos[k] = v
	}

	url = svninfos["URL"]
	for _, v := range svns() {
		if strings.HasPrefix(url, v) {
			root = v
			break
		}
	}
	if root == "" {
		fatal("需要在指定的仓库运行，请配置环境变量：JV_PATHS")
	}

	who := username()
	pattern := fmt.Sprintf(`%s/(trunk/|branches/%s/.*?/)`, root, who)
	if ok, _ := regexp.MatchString(pattern, url); ok {
		fatal("请在svn根目录执行")
	}

	trunk = fmt.Sprintf("%s/trunk", root)
	home = fmt.Sprintf("%s/branches/%s", root, who)

	if isModified() {
		fmt.Printf("\n%s\n本地有未提交的修改\n%s\n", strings.Repeat("*", 20), strings.Repeat("*", 20))
	}

	return
}

func getCmds() (cmds [][]string) {
	var trunk, home, root, url string

	switch {
	case Checkout:
		trunk := chooseTrunk()
		name := input("\n请重新命名: ")
		if name == "" {
			fatal("需要重新命名")
		}
		cmd := []string{"svn", "checkout", trunk, name}
		cmds = append(cmds, cmd)
	case Branch:
		trunk, home, root, url = getInfos()
		name := input("\n请给一个分支名:\n")
		if name == "" {
			fatal("需要分支名")
		}
		branch := fmt.Sprintf("%s/%s", home, name)
		if branchExist(branch) {
			fatal("branch already exist")
		}
		comment := input("\n请给一个描述:\n")
		if comment == "" {
			fatal("需要描述")
		}
		cmd := []string{"svn", "copy", trunk, branch, "-m", "\"" + comment + "\""}
		cmds = append(cmds, cmd)
	case Switch:
		trunk, home, root, url = getInfos()
		var branch string
		name := chooseBranch(home, true)
		if name == "/trunk" {
			branch = trunk
		} else {
			branch = fmt.Sprintf("%s/%s", home, name)
		}
		cmd := []string{"svn", "switch", "--force", branch}
		cmds = append(cmds, cmd)
	case Merge:
		trunk, home, root, url = getInfos()
		name := input("\n您想合并谁的分支代码（自己的直接回车）:\n")
		if name != "" {
			home = fmt.Sprintf("%s/branches/%s", root, name)
		}
		name = chooseBranch(home, false)
		branch := fmt.Sprintf("%s/%s", home, name)
		cmd := []string{"svn", "merge", branch}
		cmds = append(cmds, cmd)
	case Catch:
		trunk, home, root, url = getInfos()
		if trunk == url {
			fatal("只能在分支上做与trunk代码的同步")
		}
		cmds = [][]string{
			[]string{"svn", "rm", "-m", "\"del branch, ready for catching up\"", url},
			[]string{"svn", "copy", "-m", getBaseMsg(url), trunk, url},
			[]string{"svn", "switch", "--force", url},
			[]string{"svn", "merge", url + "@" + getLatestRev(url)},
		}
	case DelBranch:
		trunk, home, root, url = getInfos()
		name := chooseBranch(home, false)
		branch := fmt.Sprintf("%s/%s", home, name)
		cmd := []string{"svn", "rm", "-m", "\"delete unused branch\"", branch}
		cmds = append(cmds, cmd)
	default:
		flag.Usage()
		os.Exit(0)
	}

	return
}

func main() {
	flag.BoolVar(&Checkout, "checkout", false, "获取主干代码")
	flag.BoolVar(&Switch, "switch", false, "列出所有branch以备switch")
	flag.BoolVar(&Merge, "merge", false, "选择一个branch来merge")
	flag.BoolVar(&Catch, "catch", false, "合并trunk最新修改")
	flag.BoolVar(&Branch, "branch", false, "新建branch")
	flag.BoolVar(&DelBranch, "delbranch", false, "删除branch")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "A partner for svn command\n")
		fmt.Fprintf(os.Stderr, "version: 1.7, Created by simplejia [11/2016]\n\n")
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cmds := getCmds()
	cmdStr := ""
	for _, cmd := range cmds {
		cmdStr += strings.Join(cmd, " ") + "\n"
	}
	fmt.Printf("\n%s\n", cmdStr)
	if input("执行吗?(y确认): ") != "y" {
		fatal("abort")
	}

	fmt.Println(strings.Repeat("-", 40))

	for _, cmd := range cmds {
		fmt.Printf(">>> %s\n", strings.Join(cmd, " "))
		run(cmd[0], cmd[1:]...)
	}
}
