package input

import (
	"fmt"
)

const (
	SYS_ASCII_TAB        = 9
	SYS_ASCII_LF         = 10
	SYS_ASCII_BACK_SPACE = 127
	SYS_ASCII_ETX        = 3
)

var history_list []string
var sys_cmd_list []string

//cursor movement
func MOVE_UP(x int) {
	if x <= 0 {
		return
	}
	fmt.Printf("\033[%dA", x)
}

func MOVE_DOWN(x int) {
	if x <= 0 {
		return
	}
	fmt.Printf("\033[%dB", x)
}
func MOVE_RIGHT(y int) {
	if y <= 0 {
		return
	}
	fmt.Printf("\033[%dC", y)
}
func MOVE_LEFT(y int) {
	if y <= 0 {
		return
	}
	fmt.Printf("\033[%dD", y)
}
func MOVE_TO(x, y int) { fmt.Printf("\033[%d;%dH", (x), (y)) }
func RESET_CURSOR()    { fmt.Printf("\033[H") }
func HIDE_CURSOR()     { fmt.Printf("\033[?25l") }
func SHOW_CURSOR()     { fmt.Printf("\033[?25h") }
func SAVE_CURSOR()     { fmt.Printf("\033[s") }
func RETURN_CURSOR()   { fmt.Printf("\033[u") }

func SetSystemCommand(cmdlist []string) {
	sys_cmd_list = append(sys_cmd_list, cmdlist...)
}

func cleanCell(num int) {
	for i := 0; i < num; i++ {
		fmt.Printf(" ")
	}
}

func GetInput(echo bool, history bool) string {
	var cmd []byte
	var leftCmd, rightCmd []byte
	var historyIndex = len(history_list)
	for {
		buf, parse := getch()
		if buf == SYS_ASCII_LF {
			break
		}
		switch buf {
		case SYS_UP:
			if !echo {
				break
			}
			if historyIndex > 0 {
				historyIndex--
				MOVE_LEFT(len(leftCmd))
				cleanCell(len(leftCmd) + len(rightCmd))
				MOVE_LEFT(len(leftCmd) + len(rightCmd))
				leftCmd = []byte(history_list[historyIndex])
				rightCmd = nil
				fmt.Printf("%s", string(leftCmd))
			}
			break
		case SYS_DOWN:
			if !echo {
				break
			}
			if historyIndex < len(history_list) {
				historyIndex++
				MOVE_LEFT(len(leftCmd))
				cleanCell(len(leftCmd) + len(rightCmd))
				MOVE_LEFT(len(leftCmd) + len(rightCmd))
				if historyIndex >= len(history_list) {
					leftCmd = nil
					rightCmd = nil
				} else {
					leftCmd = []byte(history_list[historyIndex])
					rightCmd = nil
				}
				fmt.Printf("%s", string(leftCmd))
			}
			break
		case SYS_LEFT:
			if !echo {
				break
			}
			if len(leftCmd) > 0 {
				MOVE_LEFT(1)
				var tmpRight []byte
				tmpRight = append(tmpRight, leftCmd[len(leftCmd)-1])
				rightCmd = append(tmpRight, rightCmd...)
				if len(leftCmd) > 1 {
					leftCmd = leftCmd[0 : len(leftCmd)-1]
				} else {
					leftCmd = nil
				}
			}
			break
		case SYS_RIGHT:
			if !echo {
				break
			}
			if len(rightCmd) > 0 {
				MOVE_RIGHT(1)
				leftCmd = append(leftCmd, rightCmd[len(rightCmd)-1])
				rightCmd = rightCmd[1:]
			}
			break
		case SYS_PARSE:
			if echo {
				fmt.Printf("%s%s", parse, string(rightCmd))
			} else {
				for i := 0; i < len(parse)+len(rightCmd); i++ {
					fmt.Printf("*")
				}
			}
			if len(rightCmd) != 0 {
				MOVE_LEFT(len(rightCmd))
			}
			leftCmd = append(leftCmd, []byte(parse)...)
			break
		case SYS_ASCII_TAB:
			if !echo {
				break
			}
			if len(sys_cmd_list) > 0 {
				var sameCmdList []string
				for _, cmd := range sys_cmd_list {
					//fmt.Printf("\n~~~cmd = %s, leftCmd = %s~~\n", cmd, leftCmd)
					if len(leftCmd) > len(cmd) {
						continue
					}
					if string(leftCmd) == string([]byte(cmd)[0:len(leftCmd)]) {
						sameCmdList = append(sameCmdList, cmd)
					}
				}
				if len(sameCmdList) > 0 {
					MOVE_LEFT(len(leftCmd))
					leftCmd = []byte(sameCmdList[0])
					fmt.Printf("%s", string(leftCmd))
					if len(sameCmdList) > 1 {
						SAVE_CURSOR()
						var showSameCmd string
						for i, cmd := range sameCmdList {
							if i != 0 {
								showSameCmd += fmt.Sprintf("\t")
							}
							showSameCmd += fmt.Sprintf("%s", cmd)
						}
						fmt.Printf("\n%s", showSameCmd)
						RETURN_CURSOR()
						//MOVE_UP(1)
					}
				}
			}
			break
		case SYS_ASCII_BACK_SPACE:
			if len(leftCmd) > 0 {
				if len(leftCmd) > 1 {
					leftCmd = leftCmd[0 : len(leftCmd)-1]
				} else {
					leftCmd = nil
				}
				MOVE_LEFT(1)
				fmt.Printf("%s%c", string(rightCmd), ' ')
				MOVE_LEFT(len(rightCmd) + 1)
			}
			break
		default:
			if echo {
				fmt.Printf("%c%s", buf, string(rightCmd))
			} else {
				fmt.Printf("*")
			}
			if len(rightCmd) != 0 {
				MOVE_LEFT(len(rightCmd))
			}
			leftCmd = append(leftCmd, byte(buf))
		}
	}
	cmd = append(cmd, leftCmd...)
	cmd = append(cmd, rightCmd...)
	if history {
		if len(cmd) != 0 {
			if len(history_list) == 0 {
				history_list = append(history_list, string(cmd))
			} else if history_list[len(history_list)-1] != string(cmd) {
				history_list = append(history_list, string(cmd))
			}
		}
	}
	return string(cmd)
}
