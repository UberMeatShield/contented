import * as _ from 'lodash';
import { ApiDef } from './api_def';

export enum ScreenAction {
  VIEW = 'view',
  PLAY_SCREEN = 'play-screen',
}

export interface ScreenClickEvent {
  screen: Screen;
  action: ScreenAction;
  screens?: Screen[];
}

function formatSeconds(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
}
export class Screen {
  public id: number;
  public src: string;
  public idx: number;

  public content_id: number;
  public size_bytes: number;
  public content_container_id: number;
  public url: string;
  public timeSeconds: number;

  constructor(obj: any = {}) {
    this.fromJson(obj);
  }

  public fromJson(raw: any) {
    if (raw) {
      Object.assign(this, raw);
      this.timeSeconds = this.parseSecondsFromScreen() || 0;
      this.links();
    }
  }

  public links() {
    this.url = `${ApiDef.contented.screens}${this.id}`;
  }

  public parseSecondsFromScreen() {
    let screenSecondRe = new RegExp(/.*ss(\d+)\.*/);
    let timeCheck = screenSecondRe.exec(this.src);
    if (timeCheck) {
      return Number.parseInt(timeCheck[1], 10);
    }
    return 0;
  }

  public toHHMMSS() {
    return formatSeconds(this.timeSeconds);
  }
}
