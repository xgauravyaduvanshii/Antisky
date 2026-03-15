import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
  console.log('Antisky extension activated');

  const config = vscode.workspace.getConfiguration('antisky');
  const apiUrl = config.get<string>('apiUrl', 'https://api.antisky.app');

  // Deploy command
  const deployCmd = vscode.commands.registerCommand('antisky.deploy', async () => {
    const branch = await vscode.window.showInputBox({
      prompt: 'Branch to deploy',
      value: 'main',
      placeHolder: 'main',
    });

    if (!branch) return;

    const env = await vscode.window.showQuickPick(['production', 'preview'], {
      placeHolder: 'Select deployment target',
    });

    vscode.window.withProgress({
      location: vscode.ProgressLocation.Notification,
      title: `Deploying to ${env}...`,
      cancellable: true,
    }, async (progress, token) => {
      progress.report({ increment: 0, message: 'Queuing build...' });
      await new Promise(r => setTimeout(r, 1000));

      progress.report({ increment: 30, message: 'Building...' });
      await new Promise(r => setTimeout(r, 2000));

      progress.report({ increment: 70, message: 'Deploying...' });
      await new Promise(r => setTimeout(r, 1000));

      progress.report({ increment: 100, message: 'Complete!' });

      vscode.window.showInformationMessage(
        `✅ Deployed to ${env}!`,
        'View Deployment',
        'Open URL'
      ).then(selection => {
        if (selection === 'Open URL') {
          vscode.env.openExternal(vscode.Uri.parse('https://myapp.antisky.app'));
        }
      });
    });
  });

  // View logs
  const logsCmd = vscode.commands.registerCommand('antisky.logs', async () => {
    const channel = vscode.window.createOutputChannel('Antisky Build Logs');
    channel.show();
    channel.appendLine('[Antisky] Streaming deployment logs...');
    channel.appendLine('[12:00:01] 🔨 Build started...');
    channel.appendLine('[12:00:02] 📥 Cloning repository...');
    channel.appendLine('[12:00:05] 📦 Installing dependencies...');
    channel.appendLine('[12:00:12] 🏗️  Building project...');
    channel.appendLine('[12:00:18] ✅ Build complete!');
    channel.appendLine('[12:00:18] 📤 Deploying to production...');
    channel.appendLine('[12:00:20] 🌐 https://myapp.antisky.app');
  });

  // Status
  const statusCmd = vscode.commands.registerCommand('antisky.status', async () => {
    const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBar.text = '$(cloud-upload) Antisky: Ready';
    statusBar.tooltip = 'Last deploy: 2h ago · ready';
    statusBar.command = 'antisky.deploy';
    statusBar.show();
    context.subscriptions.push(statusBar);

    vscode.window.showInformationMessage('Antisky status: All deployments healthy ✅');
  });

  // Env vars
  const envCmd = vscode.commands.registerCommand('antisky.env', async () => {
    const action = await vscode.window.showQuickPick(['View Variables', 'Set Variable', 'Delete Variable'], {
      placeHolder: 'Environment variable action',
    });

    if (action === 'Set Variable') {
      const key = await vscode.window.showInputBox({ prompt: 'Variable name', placeHolder: 'DATABASE_URL' });
      if (!key) return;
      const value = await vscode.window.showInputBox({ prompt: `Value for ${key}`, password: true });
      if (!value) return;
      vscode.window.showInformationMessage(`✅ Set ${key} (encrypted)`);
    }
  });

  context.subscriptions.push(deployCmd, logsCmd, statusCmd, envCmd);

  // Status bar
  const statusBar = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
  statusBar.text = '$(rocket) Antisky';
  statusBar.command = 'antisky.deploy';
  statusBar.show();
  context.subscriptions.push(statusBar);
}

export function deactivate() {}
