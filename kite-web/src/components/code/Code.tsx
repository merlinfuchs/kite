import styles from "./Code.module.css";
import CodeEditor from "./CodeEditor";
import CodeNav from "./CodeNav";
import CodeFileTree from "./CodeFileTree";
import CodeTerminal from "./CodeTerminal";
import { FlatFile } from "@/lib/code/filetree";

interface Props {
  files: FlatFile[];
  openFilePath: string | null;
  setOpenFilePath: (path: string | null) => void;
  hasUnsavedChanges: boolean;
  onChange: () => void;
  isSaving: boolean;
  onSave: () => void;
  isDeploying: boolean;
  onDeploy: () => void;
  onExit: () => void;
}

export default function Code({
  files,
  openFilePath,
  setOpenFilePath,
  hasUnsavedChanges,
  onChange,
  isSaving,
  onSave,
  isDeploying,
  onDeploy,
  onExit,
}: Props) {
  return (
    <div className={styles.code}>
      <div className={styles.codeNav}>
        <CodeNav
          hasUnsavedChanges={hasUnsavedChanges}
          isSaving={isSaving}
          onSave={onSave}
          onExit={onExit}
          isDeploying={isDeploying}
          onDeploy={onDeploy}
        />
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
            <CodeEditor
              files={files}
              openFilePath={openFilePath}
              onChange={onChange}
              onSave={onSave}
              onDeploy={onDeploy}
            />
          </div>
          <div className={styles.codeTerminal}>
            <CodeTerminal />
          </div>
        </div>
      </div>
    </div>
  );
}
