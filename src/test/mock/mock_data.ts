// TODO: This was from before the httpMock was actually good, just use httpMockController now
import {Observable, from as observableFrom} from 'rxjs';
import {Directory} from './../../contented/directory';
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

    public getMedia(container_id = null, count = null) {
        let media = _.clone(require('./media.json'));
        if (container_id) {
            _.each(media, m => {
                m.container_id = container_id;
            });
        }
        // TODO: Create fake media / id info if given a count
        return media;
    }

    public getFullDirectory() {
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
            path: 'narp/',
            id: containerId,
            contents: contents
        };
        return fakeDirResponse;
    }

    public mockContentedService(service) {
        let previewJson = this.getPreview();
        let dirs = _.map(_.get(previewJson, 'results'), dir => {
           return new Directory(dir);
        });
        service.getPreview = this.obs(dirs);
        service.getFullDirectory = this.obs(this.getFullDirectory());
    }

    public handleCmpDefaultLoad(httpMock) {
         let containers = this.getPreview();
         let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
         containersReq.flush(containers);

         let url = ApiDef.contented.media.replace("{dirId}", '' + containers[0].id);
         let mediaReq = httpMock.expectOne(req => req.url === url);
         mediaReq.flush(MockData.getMedia());
    }


    // This will actually fake an async call to prove things require async ticks, better tests on cmps
    public obs(response, shouldReject: boolean = false) {
        let val = response;
        let timeout = this.timeoutSpan;
        return function() {
            console.log("Calling the damn method at least, promise not resolving?", timeout);
            let p = new Promise((resolve, reject) => {
                setTimeout(() => {
                    return shouldReject ? reject(val) : resolve(val);
                }, timeout);
            });
            return observableFrom(p);
        };
    }
}
export let MockData = new MockLoader();
