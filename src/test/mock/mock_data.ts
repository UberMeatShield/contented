import {Observable, from as observableFrom} from 'rxjs';
import {Container} from './../../contented/container';
import {Content} from './../../contented/content';
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

    public getVideos() {
        return _.cloneDeep(require('./video_view.json'));
    }

    public getScreens() {
        return _.clone(require('./screens.json'));
    }

    public splash() {
        return _.clone(require('./splash.json'));
    }

    public tags() {
        return _.clone(require('./tags.json'));
    }

    public videoContent() {
        return _.clone(require('./video_content.json'));
    }

    public taskRequest(taskId: string) {
        let tasks = _.clone(require('./task_requests.json'));
        let task = tasks[0];
        task.id = taskId;
        return task
    }

    public taskRequests() {
        return _.clone(require('./task_requests.json'));
    }

    public getContent(container_id = null, count = null) {
        let content = _.clone(require('./content.json'));
        if (container_id) {
            _.each(content, m => {
                m.id = m.id + container_id;
                m.container_id = container_id;
            });
        }
        // TODO: Create fake content / id info if given a count
        return content.slice(0, count);
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
        let containers = this.handleContainerLoad(httpMock)

        if (fixture) {
            fixture.detectChanges();
            this.handleContainerContentLoad(httpMock, containers);
        }
    }

    public handleContainerLoad(httpMock) {
        let containers = this.getPreview();
        let containersReq = httpMock.expectOne(req => req.url === ApiDef.contented.containers);
        containersReq.flush(containers);
        return containers;
    }

    public handleContainerContentLoad(httpMock, cnts: Array<Container>, count = 2) {
        _.each(cnts, cnt => {
            let url = ApiDef.contented.containerContent.replace('{cId}', cnt.id);
            let reqs = httpMock.match(r => r.url === url);
            _.each(reqs, req => {
                req.flush(MockData.getContent(cnt.name, count));
            });
        });
    }

    public getImg() {
        let img = new Content();
        img.fromJson(this.getContent("10", 1)[0]);
        return img;
    }
}
export let MockData = new MockLoader();
