import {Observable} from 'rxjs';
class MockLoader {

    public constructor() {

    }

    public getPreview() {
        return require('preview.json');
    }

    public getFullDirectory() {
        return require('./full_dir.json');
    }

    public mockContentedService(service) {
        service.getPreview = this.obs(this.getPreview());
        service.getFullDirectory = this.obs(this.getFullDirectory());
    }

    // This will actually fake an async call to prove things require async ticks, better tests on cmps
    public timeoutSpan = 100;
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
            return Observable.fromPromise(p);
        };
    }
}

export let MockData = new MockLoader();
