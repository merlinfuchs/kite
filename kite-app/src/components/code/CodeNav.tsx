import {
  ArrowLeftIcon,
  ArrowPathIcon,
  ArrowUpIcon,
  CheckIcon,
} from "@heroicons/react/20/solid";
import styles from "./CodeNav.module.css";
import clsx from "clsx";

interface Props {
  hasUnsavedChanges: boolean;
  isSaving: boolean;
  onBack(): void;
  onSave(): void;
  isDeploying: boolean;
  onDeploy(): void;
}

export default function CodeNav({
  hasUnsavedChanges,
  isSaving,
  onBack,
  onSave,
  isDeploying,
  onDeploy,
}: Props) {
  return (
    <div className={styles.root}>
      <button className={styles.item} onClick={onBack}>
        <ArrowLeftIcon className={styles.icon} />
        <div>Back to Server</div>
      </button>
      {isSaving ? (
        <div className={styles.item} onClick={onSave}>
          <ArrowPathIcon className={clsx(styles.icon, styles.spin)} />
          Saving Changes
        </div>
      ) : hasUnsavedChanges ? (
        <button className={styles.item} onClick={onSave}>
          <div className={styles.dot} />
          Save Changes
        </button>
      ) : (
        <div className={styles.item}>
          <CheckIcon className={styles.icon} />
          No Unsaved Changes
        </div>
      )}
      {isDeploying ? (
        <div className={styles.item} onClick={onSave}>
          <ArrowPathIcon className={clsx(styles.icon, styles.spin)} />
          Deploying to Server
        </div>
      ) : (
        <button className={styles.item} onClick={onDeploy}>
          <ArrowUpIcon className={styles.icon} />
          Deploy to Server
        </button>
      )}
    </div>
  );
}
