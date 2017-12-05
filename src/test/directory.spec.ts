import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';

import * as _ from 'lodash';
import {Directory} from './../contented/directory';
import {MockData} from './mock/mock_data';

describe('TestingDirectory', () => {

    it('Should be able to create a directory.', () => {
        let d = new Directory({});
    });

    it('Should be able to create a set of directory objects', () => {
        let dirResponse = MockData.getPreview();
        let dirs = _.map(_.get(dirResponse, 'results'), data => {
            return new Directory(data);
        });
        expect(dirs.length > 0).toBe(true, "It should actually have some responses.");
        _.each(dirs, dir => {
            expect(dir.getContentList().length > 0).toBe(true, "We should have a content list and be able to build them out.");
            expect(dir.id).toBeDefined("We should have an id set for each dir.");
        });
    });

    function getMockDir(count: number) {
        let total = 20;
        let fakeDirResponse = {
            total: total,
            path: 'narp/',
            id: 'test',
            contents: _.map(_.range(0, count), idx => 'item-' + idx)
        };
        return fakeDirResponse;
    }

    it('Should be able to setup intervals successfully', () => {
        let total = 20;
        let dir = new Directory(getMockDir(total));
        let contents = dir.getContentList();
        expect(contents.length).toBe(total, "We should have an entry for each item");

        let testIdx = 5;
        let interval = dir.getIntervalAround(contents[testIdx], 5, 1);
        expect(interval.length).toBe(5, "We should get a 3 item interval");

        let targetIdx = _.indexOf(interval, contents[testIdx - 1]);
        expect(targetIdx).toBe(0, "It should be in the first result (the previous item)");
        expect(_.indexOf(interval, contents[testIdx + 1])).toBe(2, "Should be the next item in the list");
        expect(_.indexOf(interval, contents[testIdx - 2])).toBe(-1, "We should not have more than 1 item before the selected item");
    });

    it('Should manage to render the requested number each time, and best effort otherwise', () => {
        let dir = new Directory(getMockDir(6));
        let items = dir.getContentList();
        let item = items[1];
        let results = dir.getIntervalAround(item, 3, 1);
        expect(results.length).toBe(3, "We should have 3 items with this selection");

        let finalItemTest = items[items.length - 1];
        let enoughResults = dir.getIntervalAround(finalItemTest, 4, 1);
        expect(enoughResults.length).toBe(4, "We should still have 3 results, (it should adjust the start)");
    });

});

