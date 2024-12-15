import * as _ from 'lodash';
import { ApiDef } from './api_def';

export class Screen {
  public id: string;
  public src: string;
  public idx: number;

  public content_id: string;
  public size_bytes: number;
  public content_container_id: string;
  public url: string;
  public timeSeconds: number;

  constructor(obj: any = {}) {
    this.fromJson(obj);
  }

  public fromJson(raw: any) {
    if (raw) {
      Object.assign(this, raw);
      this.timeSeconds = this.parseSecondsFromScreen();
      this.links();
    }
  }

  public links() {
    this.url = `${ApiDef.contented.screens}${this.id}`;
  }

  public parseSecondsFromScreen() {
    let screenSecondRe = new RegExp(/.*ss(\d+)\.jpg/);
    let timeCheck = screenSecondRe.exec(this.src);
    if (timeCheck) {
      return Number.parseInt(timeCheck[1], 10);
    }
    return 0;
  }
}
