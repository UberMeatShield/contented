import {Observable, from as observableFrom} from 'rxjs';
import {Container} from './../../contented/container';
import {Media} from './../../contented/media';
import {ApiDef} from './../../contented/api_def';
import * as _ from 'lodash';

declare var require: any;
class MockLoader {

    public timeoutSpan = 100;
    public constructor() {

    }

    public getPreview() {
        return _.clone(require('./containers.json'));
    }

    public getSearch() {
        return _.clone(require('./search.json'));
    }

    public getScreens() {
        return _.clone(require('./screens.json'));
    }

    public getMedia(container_id = null, count = null) {
        let media = _.clone(require('./media.json'));
        if (container_id) {
            _.each(media, m => {
                m.id = m.id + container_id;
                m.container_id = container_id;
            });
        }
        // TODO: Create fake media / id info if given a count
        return media.slice(0, count);
    }

    public getFullContainer() {
        return require('./full.json');
    }

    public getMockDir(count: number, itemPrefix: string = 'item-', offset: number = 0, total = 20) {
        let containerId = 'test';
        let contents = _.map(_.range(0, count),
            (idx) => {
                let id = idx + offset;
                return {src: itemPrefix + id, id: id, container_id: containerId};
            }
        );

        let fakeDirResponse = {
            total: total,
            path: `narp${containerId}/`,
            name: containerId,  // Generate a UUID?
            id: containerId,
            contents: contents  // Note the API does not currently return contents
        };
        return fakeDirResponse;
    }

    public handleCmpDefaultLoad(httpMock, fixture = null) {
         let containers = this.getPreview();
         let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
         containersReq.flush(containers);

        if (fixture) {
            fixture.detectChanges();
            this.handleContainerMediaLoad(httpMock, containers);
        }
    }

    public handleContainerMediaLoad(httpMock, cnts: Array<Container>, count = 2) {
        _.each(cnts, cnt => {
            let url = ApiDef.contented.media.replace('{cId}', cnt.id);
            let reqs = httpMock.match(r => r.url === url);
            _.each(reqs, req => {
                req.flush(MockData.getMedia(cnt.name, count));
            });
        });
    }

    public getImg() {
        let img = new Media();
        img.fromJson(this.getMedia("10", 1)[0]);
        return img;
    }
}
export let MockData = new MockLoader();
