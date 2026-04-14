package main

import (
    "os"
    "path/filepath"
    "testing"
)

func TestBuildDesktopINIContent(t *testing.T) {
    exePath := `C:\Tools\Demo App.exe`

    got := buildDesktopINIContent("Demo App", exePath)
    want := "[.ShellClassInfo]\r\n" +
        "LocalizedResourceName=Demo App\r\n" +
        "IconResource=C:\\Tools\\Demo App.exe,0\r\n" +
        "[ViewState]\r\n" +
        "Mode=\r\n" +
        "Vid=\r\n" +
        "FolderType=Generic\r\n"

    if got != want {
        t.Fatalf("unexpected desktop.ini content\nwant:\n%q\n\ngot:\n%q", want, got)
    }
}

func TestProcessExeDropCreatesDesktopINI(t *testing.T) {
    tempDir := t.TempDir()
    exePath := filepath.Join(tempDir, "Demo.exe")
    if err := os.WriteFile(exePath, []byte("binary"), 0o644); err != nil {
        t.Fatalf("write exe: %v", err)
    }

    if err := processExeDrop(exePath); err != nil {
        t.Fatalf("process exe drop: %v", err)
    }

    iniPath := filepath.Join(tempDir, "desktop.ini")
    data, err := os.ReadFile(iniPath)
    if err != nil {
        t.Fatalf("read desktop.ini: %v", err)
    }

    want := buildDesktopINIContent("Demo", exePath)
    if string(data) != want {
        t.Fatalf("desktop.ini mismatch\nwant:\n%q\n\ngot:\n%q", want, string(data))
    }
}

func TestProcessFolderDropCreatesDesktopINI(t *testing.T) {
    tempDir := t.TempDir()
    folderPath := filepath.Join(tempDir, "Games")
    iconPath := filepath.Join(tempDir, "launcher.exe")

    if err := os.Mkdir(folderPath, 0o755); err != nil {
        t.Fatalf("mkdir folder: %v", err)
    }
    if err := os.WriteFile(iconPath, []byte("binary"), 0o644); err != nil {
        t.Fatalf("write icon exe: %v", err)
    }

    if err := processFolderDrop(folderPath, "My Games", iconPath); err != nil {
        t.Fatalf("process folder drop: %v", err)
    }

    iniPath := filepath.Join(folderPath, "desktop.ini")
    data, err := os.ReadFile(iniPath)
    if err != nil {
        t.Fatalf("read desktop.ini: %v", err)
    }

    want := buildDesktopINIContent("My Games", iconPath)
    if string(data) != want {
        t.Fatalf("desktop.ini mismatch\nwant:\n%q\n\ngot:\n%q", want, string(data))
    }
}
