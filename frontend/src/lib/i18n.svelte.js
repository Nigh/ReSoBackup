const translations = {
  en: {
    appTitle: 'Reed-Solomon Backup',
    backupTab: 'Backup',
    restoreTab: 'Restore',
    footer: 'Reed-Solomon Encrypted Backup Tool · AES-256-GCM + scrypt KDF',
    inputFile: 'Input File',
    browse: 'Browse',
    selectFile: 'Select a file to backup...',
    rsParams: 'Reed-Solomon Parameters',
    totalShares: 'Total Shares',
    threshold: 'Threshold (min to restore)',
    redundancy: 'Redundancy',
    canLose: '{n} share(s) can be lost',
    storageOverhead: 'Storage Overhead',
    sharesTotal: '{n} shares total',
    encryption: 'Encryption',
    password: 'Password',
    enterPassword: 'Enter backup password...',
    outputDir: 'Output Directory',
    outputDirPlaceholder: 'Same as input file directory (default)',
    outputDirHint: 'Leave empty to use the same directory as the input file',
    createBackup: 'Create Backup',
    creatingBackup: 'Creating Backup...',
    backupSuccess: 'Backup completed successfully!',
    shareFile: 'Share File or Metadata',
    selectRestore: 'Select any .rs.NNN or .rsmeta file...',
    restoreHint: 'You can select any share file (.rs.001, .rs.002, ...) or the metadata file (.rsmeta). All available shares in the same directory will be used for reconstruction.',
    decryptPwd: 'Decryption Password',
    enterRestorePwd: 'Enter the backup password...',
    restoreOutputPlaceholder: 'Current directory (default)',
    restoreOutputHint: 'Leave empty to restore to the current working directory',
    restoreFile: 'Restore File',
    restoring: 'Restoring...',
    restoreSuccess: 'File restored successfully!',
    errSelectFile: 'Please select a file to backup',
    errEnterPwd: 'Please enter a password',
    errSelectRestore: 'Please select a share file or .rsmeta file',
    errEnterRestorePwd: 'Please enter the backup password',
    errPrefix: 'error',
    warnTitle: 'Warnings detected:',
    confirmContinue: 'Do you want to continue?',
    sizeUnit: 'Size',
  },
  zh: {
    appTitle: 'Reed-Solomon 备份工具',
    backupTab: '备份',
    restoreTab: '恢复',
    footer: 'Reed-Solomon 加密备份工具 · AES-256-GCM + scrypt KDF',
    inputFile: '输入文件',
    browse: '浏览',
    selectFile: '选择要备份的文件...',
    rsParams: 'Reed-Solomon 参数',
    totalShares: '总分片数',
    threshold: '恢复阈值（最少需要）',
    redundancy: '冗余度',
    canLose: '可丢失 {n} 个分片',
    storageOverhead: '存储开销',
    sharesTotal: '共 {n} 个分片',
    encryption: '加密',
    password: '密码',
    enterPassword: '输入备份密码...',
    outputDir: '输出目录',
    outputDirPlaceholder: '与输入文件同目录（默认）',
    outputDirHint: '留空则输出到输入文件所在目录',
    createBackup: '创建备份',
    creatingBackup: '正在创建备份...',
    backupSuccess: '备份创建成功！',
    shareFile: '分片文件或元数据',
    selectRestore: '选择任意 .rs.NNN 或 .rsmeta 文件...',
    restoreHint: '可以选择任意分片文件（.rs.001, .rs.002, ...）或元数据文件（.rsmeta）。同目录下所有可用分片将用于重建。',
    decryptPwd: '解密密码',
    enterRestorePwd: '输入备份密码...',
    restoreOutputPlaceholder: '当前目录（默认）',
    restoreOutputHint: '留空则恢复到当前工作目录',
    restoreFile: '恢复文件',
    restoring: '正在恢复...',
    restoreSuccess: '文件恢复成功！',
    errSelectFile: '请选择要备份的文件',
    errEnterPwd: '请输入密码',
    errSelectRestore: '请选择分片文件或 .rsmeta 文件',
    errEnterRestorePwd: '请输入备份密码',
    errPrefix: '错误',
    warnTitle: '检测到警告：',
    confirmContinue: '是否继续？',
    sizeUnit: '大小',
  }
};

let currentLang = $state('en');
let currentT = $state(translations.en);

export function getLang() {
  return currentLang;
}

export function getT() {
  return currentT;
}

export function setLang(l) {
  currentLang = l;
  currentT = translations[l];
}

export function toggleLang() {
  currentLang = currentLang === 'en' ? 'zh' : 'en';
  currentT = translations[currentLang];
}
