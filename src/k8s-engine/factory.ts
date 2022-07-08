import os from 'os';

import { Architecture, KubernetesBackend } from './k8s';
import MockBackend from './mock';
import LimaBackend from './lima';
import WSLBackend from './wsl';
import DockerDirManager from '@/utils/dockerDirManager';

export default function factory(arch: Architecture, dockerDirManager: DockerDirManager): KubernetesBackend {
  const platform = os.platform();

  if (process.env.RD_MOCK_BACKEND === '1') {
    return new MockBackend();
  }

  switch (platform) {
  case 'linux':
    return new LimaBackend(arch, dockerDirManager);
  case 'darwin':
    return new LimaBackend(arch, dockerDirManager);
  case 'win32':
    return new WSLBackend();
  default:
    throw new Error(`OS "${ platform }" is not supported.`);
  }
}
