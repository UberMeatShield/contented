import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import {ContentedCmp} from '../contented/contented_cmp';
import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingContentedCmp', () => {
    let fixture: ComponentFixture<ContentedCmp>;
    let service: ContentedService;
    let comp: ContentedCmp;
    let el: HTMLElement;
    let de: DebugElement;

    beforeEach(async( () => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule, ContentedModule, HttpClientTestingModule],
            providers: [
                ContentedService
            ]
        }).compileComponents();

        service = TestBed.get(ContentedService);
        fixture = TestBed.createComponent(ContentedCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-cmp'));
        el = de.nativeElement;
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        MockData.mockContentedService(comp._contentedService);
        fixture.detectChanges();
        tick(2000);
        expect(comp.allD.length).toBe(3, "We should have 3 directories set");

        let dirs = comp.getVisibleDirectories();
        expect(dirs.length).toBe(comp.maxVisible, "Should only have the max visible directories present.");
        expect(dirs.length <= comp.allD.length).toBe(true, "It should never have more data than we asked for.");

        fixture.detectChanges();
        let dirEls = $('.directory-contents', el);
        expect(dirEls.length).toBe(comp.maxVisible, "We should have the elements rendered.");

        expect($('.current-content-dir').length).toBe(1, "We should only have 1 selected dir");
    }));


    it('Should handle a click event to show a particular image.', fakeAsync(() => {
        MockData.mockContentedService(comp._contentedService);
        fixture.detectChanges();
        tick(2000);

        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode");
        expect($('.contented-view-cmp').length).toBe(0, "It should now have a view component.");

        fixture.detectChanges();
        let imgs = $('.preview-img');
        expect(imgs.length > 1).toBe(true, "A bunch of images should be visible");
        expect(comp.fullScreen).toBe(false, "We should not be in fullsceen mode even after everything is loaded");

        let toClick = $(imgs[3]).trigger('click');
        expect(comp.fullScreen).toBe(true, "It should now have a selected item");
        expect(comp.getCurrentLocation()).toBe(imgs[3].src, "It should have the current item as the image");
    }));
});

