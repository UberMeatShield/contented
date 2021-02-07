// TODO: This was from before the httpMock was actually good, just use httpMockController now
import {Observable, from as observableFrom} from 'rxjs';
import {Directory} from './../../contented/directory';
import * as _ from 'lodash';

declare var require: any;
class MockLoader {

    public timeoutSpan = 100;
    public constructor() {

    }

    public getPreview() {
        return require('./containers.json');
    }

    public getFullDirectory() {
        return require('./full.json');
    }

    public getMockDir(count: number, itemPrefix: string = 'item-', offset: number = 0, total = 20) {
         let contents = _.map(_.range(0, count),
             (idx) => {
                 let id = idx + offset;
                 return {src: itemPrefix + id, id: id};
             }
         );

         let fakeDirResponse = {
             total: total,
             path: 'narp/',
             id: 'test',
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
