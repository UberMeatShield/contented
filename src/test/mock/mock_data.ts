// TODO: This was from before the httpMock was actually good, just use httpMockController now
import {Observable, from as observableFrom} from 'rxjs';
import * as _ from 'lodash';

declare var require: any;
class MockLoader {

    public timeoutSpan = 100;
    public constructor() {

    }

    public getPreview() {
        return require('./preview.json');
    }

    public getFullDirectory() {
        return require('./full.json');
    }

    public getMockDir(count: number, itemPrefix: string = 'item-') {
         let total = 20;
         let fakeDirResponse = {
             total: total,
             path: 'narp/',
             id: 'test',
             contents: _.map(_.range(0, count), idx => itemPrefix + idx)
         };
         return fakeDirResponse;
     }

    public mockContentedService(service) {
        service.getPreview = this.obs(this.getPreview());
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
