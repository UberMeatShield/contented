import * as _ from 'lodash';
import { ApiDef } from './api_def';

import { z } from 'zod';
import { Z } from 'zod-class';

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
export class Screen extends Z.class({
  id: z.number(),
  src: z.string(),
  idx: z.number().default(0),
  content_id: z.number(),
  size_bytes: z.number().default(0),
  content_container_id: z.number().optional(),
}) {
  get timeSeconds() {
    return this.parseSecondsFromScreen() || 0;
  }

  get url() {
    return `${ApiDef.contented.screens}${this.id}`;
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
