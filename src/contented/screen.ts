import * as _ from 'lodash';
import {ApiDef} from './api_def';

export class Screen {
    public id: string;
    public src: string;
    public idx: number;

    public media_container_id: string;
    public url: string;

    constructor(obj: any = {}) {
        this.fromJson(obj);
    }

    public fromJson(raw: any) {
        if (raw) {
            Object.assign(this, raw);
            this.links();
        }
    }

    public links() {
        this.url = `${ApiDef.contented.screens}${this.id}`;
    }
}
