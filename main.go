package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
    if len(os.Args) < 2 {
        showInfo("使用说明", "把 exe 或文件夹拖到这个工具上即可。")
        return
    }

    for _, targetPath := range os.Args[1:] {
        if err := processDropTarget(targetPath); err != nil {
            showError("处理失败", err)
            return
        }
    }

    showInfo("处理完成", "desktop.ini 已生成。")
}

func processDropTarget(targetPath string) error {
    cleanedPath := strings.TrimSpace(targetPath)
    if cleanedPath == "" {
        return errors.New("收到空路径")
    }

    info, err := os.Stat(cleanedPath)
    if err != nil {
        return fmt.Errorf("读取路径失败: %w", err)
    }

    if info.IsDir() {
        return handleFolderDrop(cleanedPath)
    }

    if !strings.EqualFold(filepath.Ext(cleanedPath), ".exe") {
        return fmt.Errorf("只支持 exe 或文件夹: %s", cleanedPath)
    }

    return processExeDrop(cleanedPath)
}

func processExeDrop(exePath string) error {
    exeName := strings.TrimSuffix(filepath.Base(exePath), filepath.Ext(exePath))
    targetDir := filepath.Dir(exePath)
    return writeDesktopINI(targetDir, exeName, exePath)
}

func processFolderDrop(folderPath string, localizedName string, iconExePath string) error {
    if strings.TrimSpace(localizedName) == "" {
        return errors.New("文件夹显示名称不能为空")
    }
    if !strings.EqualFold(filepath.Ext(iconExePath), ".exe") {
        return fmt.Errorf("图标文件必须是 exe: %s", iconExePath)
    }

    if info, err := os.Stat(iconExePath); err != nil {
        return fmt.Errorf("读取图标 exe 失败: %w", err)
    } else if info.IsDir() {
        return fmt.Errorf("图标路径不能是文件夹: %s", iconExePath)
    }

    return writeDesktopINI(folderPath, localizedName, iconExePath)
}

func handleFolderDrop(folderPath string) error {
    defaultName := filepath.Base(folderPath)
    localizedName, err := promptFolderName(defaultName)
    if err != nil {
        return err
    }

    iconPath, err := promptIconExecutable(folderPath)
    if err != nil {
        return err
    }

    return processFolderDrop(folderPath, localizedName, iconPath)
}

func writeDesktopINI(targetDir string, localizedName string, iconExePath string) error {
    if info, err := os.Stat(targetDir); err != nil {
        return fmt.Errorf("读取目标目录失败: %w", err)
    } else if !info.IsDir() {
        return fmt.Errorf("目标不是文件夹: %s", targetDir)
    }

    iniPath := filepath.Join(targetDir, "desktop.ini")
    content := buildDesktopINIContent(localizedName, iconExePath)
    if err := os.WriteFile(iniPath, []byte(content), 0o644); err != nil {
        return fmt.Errorf("写入 desktop.ini 失败: %w", err)
    }

    if err := markDesktopINIAttributes(iniPath); err != nil {
        return err
    }
    if err := markFolderCustomizable(targetDir); err != nil {
        return err
    }

    return nil
}

func buildDesktopINIContent(localizedName string, iconExePath string) string {
    return fmt.Sprintf("[.ShellClassInfo]\r\nLocalizedResourceName=%s\r\nIconResource=%s,0\r\n[ViewState]\r\nMode=\r\nVid=\r\nFolderType=Generic\r\n", localizedName, iconExePath)
}

func promptFolderName(defaultName string) (string, error) {
    promptScript := `Add-Type -AssemblyName Microsoft.VisualBasic;$value=[Microsoft.VisualBasic.Interaction]::InputBox('请输入文件夹显示名称','文件夹名称','` + escapeForSingleQuotedPowerShell(defaultName) + `');if([string]::IsNullOrWhiteSpace($value)){exit 2};[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;Write-Output $value`
    output, err := runPowerShell(promptScript)
    if err != nil {
        return "", err
    }

    name := strings.TrimSpace(output)
    if name == "" {
        return "", errors.New("文件夹显示名称不能为空")
    }
    return name, nil
}

func promptIconExecutable(startDir string) (string, error) {
    dialogScript := `$dialog=New-Object System.Windows.Forms.OpenFileDialog;$dialog.Title='选择用作文件夹图标的 exe';$dialog.Filter='Executable (*.exe)|*.exe';$dialog.InitialDirectory='` + escapeForSingleQuotedPowerShell(startDir) + `';if($dialog.ShowDialog() -ne [System.Windows.Forms.DialogResult]::OK){exit 2};[Console]::OutputEncoding=[System.Text.Encoding]::UTF8;Write-Output $dialog.FileName`
    output, err := runPowerShellWithForms(dialogScript)
    if err != nil {
        return "", err
    }

    iconPath := strings.TrimSpace(output)
    if iconPath == "" {
        return "", errors.New("未选择图标 exe")
    }
    return iconPath, nil
}

func markDesktopINIAttributes(iniPath string) error {
    if err := runAttrib("+h", "+s", iniPath); err != nil {
        return fmt.Errorf("设置 desktop.ini 属性失败: %w", err)
    }
    return nil
}

func markFolderCustomizable(folderPath string) error {
    if err := runAttrib("+r", folderPath); err != nil {
        return fmt.Errorf("设置文件夹属性失败: %w", err)
    }
    return nil
}

func runAttrib(args ...string) error {
    cmd := exec.Command("attrib", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("attrib 执行失败: %s", strings.TrimSpace(string(output)))
    }
    return nil
}

func runPowerShell(script string) (string, error) {
    cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
    output, err := cmd.CombinedOutput()
    if err == nil {
        return string(output), nil
    }

    if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 2 {
        return "", errors.New("用户已取消")
    }

    return "", fmt.Errorf("PowerShell 执行失败: %s", strings.TrimSpace(string(output)))
}

func runPowerShellWithForms(script string) (string, error) {
    formsScript := "Add-Type -AssemblyName System.Windows.Forms;" + script
    return runPowerShell(formsScript)
}

func showInfo(title string, message string) {
    _ = showMessageBox(title, message, "Information")
}

func showError(title string, err error) {
    _ = showMessageBox(title, err.Error(), "Error")
}

func showMessageBox(title string, message string, icon string) error {
    script := `Add-Type -AssemblyName PresentationFramework;[System.Windows.MessageBox]::Show('` + escapeForSingleQuotedPowerShell(message) + `','` + escapeForSingleQuotedPowerShell(title) + `','OK','` + escapeForSingleQuotedPowerShell(icon) + `') | Out-Null`
    _, err := runPowerShell(script)
    return err
}

func escapeForSingleQuotedPowerShell(value string) string {
    return strings.ReplaceAll(value, `'`, `''`)
}