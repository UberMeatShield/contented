import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';

import * as _ from 'lodash';
import { Container } from './../contented/container';
import { Content } from './../contented/content';
import { MockData } from '../test/mock/mock_data';

describe('TestingContainer', () => {
  it('Should be able to create a container.', () => {
    let d = new Container({ id: 0 });
  });

  it('Should be able to create a set of container objects', () => {
    let dirResponse = MockData.getPreview();
    let cnts = _.map(dirResponse.results, data => {
      return new Container(data);
    });
    expect(cnts.length > 0).toBe(true, 'It should actually have some responses.');
    _.each(cnts, dir => {
      expect(dir.total).toBeGreaterThan(0, 'There should be contents');
      expect(dir.id).toBeDefined('We should have an id set for each dir.');
    });
  });

  it('Should be able to setup intervals successfully', () => {
    let total = 20;
    let dir = new Container(MockData.getMockDir(total));
    let contents = dir.getContentList();
    expect(contents.length).toBe(total, 'We should have an entry for each item');

    let testIdx = 5;
    let interval = dir.getIntervalAround(contents[testIdx], 5, 1);
    expect(interval.length).toBe(5, 'We should get a 3 item interval');

    let targetIdx = dir.indexOf(contents[testIdx - 1], interval);
    expect(targetIdx).toBe(0, 'It should be in the first result (the previous item)');

    expect(dir.indexOf(contents[testIdx + 1], interval)).toBe(2, 'Should be the next item in the list');
    expect(dir.indexOf(contents[testIdx - 2], interval)).toBe(
      -1,
      'We should not have more than 1 item before the selected item'
    );
  });

  it('Should manage to render the requested number each time, and best effort otherwise', () => {
    let dir = new Container(MockData.getMockDir(6));
    let items = dir.getContentList();
    let item = items[1];
    let results = dir.getIntervalAround(item, 3, 1);
    expect(results.length).toBe(3, 'We should have 3 items with this selection');

    let finalItemTest = items[items.length - 1];
    let enoughResults = dir.getIntervalAround(finalItemTest, 4, 1);
    expect(enoughResults.length).toBe(4, 'We should still have 3 results, (it should adjust the start)');
  });

  it('Should add more content with more data', () => {
    let dir = new Container(MockData.getMockDir(0));
    dir.setContents(
      [
        { id: 0, src: 'a' },
        { id: 1, src: 'b' },
      ].map(c => new Content(c))
    );

    dir.total = 5;
    expect(dir.count).toBe(2, 'We should have 2 results');
    dir.addContents(
      [
        { id: 2, src: 'c' },
        { id: 0, src: 'a' },
        { id: 3, src: 'd' },
      ].map(c => new Content(c))
    );
    expect(dir.contents.length).toBe(4, 'It should have added contents');
    expect(dir.count).toBe(4, 'The count should have updated');
  });
});
