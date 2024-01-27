export interface File {
  path: string;
  name: string;
  content: string;
}

export interface Directory {
  path: string;
  name: string;
  files: File[];
  directories: Directory[];
}

export interface FlatFile {
  path: string;
  content: string;
}

export function flatFilesToFileTree(files: FlatFile[]) {
  const root: Directory = {
    path: "",
    name: "",
    files: [],
    directories: [],
  };

  for (const file of files) {
    const path = file.path.split("/");
    let currentDir = root;

    for (let i = 0; i < path.length - 1; i++) {
      const dirName = path[i];

      let dir = currentDir.directories.find((d) => d.name === dirName);
      if (!dir) {
        dir = {
          path: path.slice(0, i + 1).join("/"),
          name: dirName,
          files: [],
          directories: [],
        };
        currentDir.directories.push(dir);
      }

      currentDir = dir;
    }

    currentDir.files.push({
      path: file.path,
      name: path[path.length - 1],
      content: file.content,
    });
  }

  return root;
}

export function getIconUrlForFile(filename: string) {
  const extension = filename.split(".").pop() || "";

  let icon = "default_file";
  switch (extension) {
    case "ts":
      icon = "file_type_typescript";
      break;
    case "js":
      icon = "file_type_js";
      break;
    case "json":
      icon = "file_type_json";
      break;
    case "yaml":
      icon = "file_type_yaml";
      break;
    case "toml":
      icon = "file_type_toml";
      break;
  }

  return `/app/file-icons/${icon}.svg`;
}
