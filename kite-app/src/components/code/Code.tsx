import styles from "./Code.module.css";
import CodeEditor from "./CodeEditor";
import CodeNav from "./CodeNav";
import CodeFileTree from "./CodeFileTree";
import CodeTerminal from "./CodeTerminal";
import { FlatFile } from "@/util/filetree";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
}

export default function Code({ files, openFilePath, setOpenFilePath }: Props) {
  return (
    <div className={styles.code}>
      <div className={styles.codeNav}>
        <CodeNav />
      </div>
      <div className={styles.codeMain}>
        <div className={styles.codeFileTree}>
          <CodeFileTree
            files={files}
            openFilePath={openFilePath}
            setOpenFilePath={setOpenFilePath}
          />
        </div>
        <div className={styles.codeCenter}>
          <div className={styles.codeEditor}>
            <CodeEditor files={files} openFilePath={openFilePath} />
          </div>
          <div className={styles.codeTerminal}>
            <CodeTerminal />
          </div>
        </div>
      </div>
    </div>
  );
}
