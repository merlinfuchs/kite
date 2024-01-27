import {
  Directory,
  File,
  FlatFile,
  flatFilesToFileTree,
  getIconUrlForFile,
} from "@/util/filetree";
import { useState } from "react";
import styles from "./CodeFileTree.module.css";
import clsx from "clsx";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
}

export default function CodeFileTree({
  files,
  openFilePath,
  setOpenFilePath,
}: Props) {
  const fileTree = flatFilesToFileTree(files);

  return (
    <div className={styles.root}>
      <div className={styles.main}>
        {fileTree.directories.map((directory) => (
          <FileTreeDirectory
            key={directory.path}
            directory={directory}
            depth={0}
            openFilePath={openFilePath}
            setOpenFilePath={setOpenFilePath}
          />
        ))}
        {fileTree.files.map((file) => (
          <FileTreeFile
            key={file.path}
            file={file}
            depth={0}
            openFilePath={openFilePath}
            setOpenFilePath={setOpenFilePath}
          />
        ))}
      </div>
      <div className={styles.divider}></div>
    </div>
  );
}

interface FileProps {
  file: File;
  depth: number;
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
}

function FileTreeFile({
  file,
  depth,
  openFilePath,
  setOpenFilePath,
}: FileProps) {
  const isOpen = file.path === openFilePath;
  const icon = getIconUrlForFile(file.name);

  return (
    <div
      className={clsx(styles.item, styles.file, isOpen && styles.open)}
      role="button"
      onClick={() => setOpenFilePath(file.path)}
    >
      <div className={styles.inner} style={{ paddingLeft: `${depth * 16}px` }}>
        <img src={icon} alt="" className={styles.icon} />
        <div className="select-none">{file.name}</div>
      </div>
    </div>
  );
}

interface DirectoryProps {
  directory: Directory;
  depth: number;
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
}

function FileTreeDirectory({
  directory,
  depth,
  openFilePath,
  setOpenFilePath,
}: DirectoryProps) {
  const [collapsed, setCollapsed] = useState(false);

  const icon = collapsed ? "default_folder.svg" : "default_folder_opened.svg";

  return (
    <div>
      <div
        className={clsx(styles.item, styles.folder)}
        onClick={() => setCollapsed(!collapsed)}
        role="button"
      >
        <div
          className={styles.inner}
          style={{ paddingLeft: `${depth * 16}px` }}
        >
          <img src={`/app/file-icons/${icon}`} alt="" className={styles.icon} />
          <div>{directory.name}</div>
        </div>
      </div>
      {!collapsed && (
        <div className={styles.children}>
          {directory.directories.map((directory) => (
            <FileTreeDirectory
              key={directory.path}
              directory={directory}
              depth={depth + 1}
              openFilePath={openFilePath}
              setOpenFilePath={setOpenFilePath}
            />
          ))}
          {directory.files.map((file) => (
            <FileTreeFile
              key={file.path}
              file={file}
              depth={depth + 1}
              openFilePath={openFilePath}
              setOpenFilePath={setOpenFilePath}
            />
          ))}
        </div>
      )}
    </div>
  );
}
