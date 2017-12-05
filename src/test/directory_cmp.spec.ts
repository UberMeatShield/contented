import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {HttpClientTestingModule} from '@angular/common/http/testing';
import {DebugElement} from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';
import {DirectoryCmp} from '../contented/directory_cmp';
import {Directory} from '../contented/directory';

import {ContentedService} from '../contented/contented_service';
import {ContentedModule} from '../contented/contented_module';

import * as _ from 'lodash';
import {MockData} from './mock/mock_data';

declare var $;
describe('TestingContentedsCmp', () => {
    let fixture: ComponentFixture<DirectoryCmp>;
    let service: ContentedService;
    let comp: DirectoryCmp;
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
        fixture = TestBed.createComponent(DirectoryCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.directory-cmp'));
        el = de.nativeElement;
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });

    it('Should be able to load up the basic data and render an image for each', fakeAsync(() => {
        let fullResponse = MockData.getFullDirectory();
        let dir = new Directory(fullResponse);

        comp.maxRendered = 3;
        comp.maxPrevItems = 0;
        comp.dir = dir;
        fixture.detectChanges();
        expect($('.preview-img', el).length).toBe(comp.maxRendered, "We should at max have items visible = " + comp.maxRendered);
    }));

    it('Should be able to page through to more items', () => {
        let fullResponse = MockData.getFullDirectory();
        let dir = new Directory(fullResponse);

        let items = dir.getContentList();
        comp.currentViewItem = items[1];
        comp.maxRendered = 4;
        comp.maxPrevItems = 1;

        // Check to ensure everything is rendering
        comp.dir = dir;
        fixture.detectChanges();
        expect($('.preview-img', el).length).toBe(comp.maxRendered, "Should select second image");

        // Now test that when we are on the last image it properly selects that
        comp.currentViewItem = items[items.length - 1]; // Choose the last item in the list.
        fixture.detectChanges();
        expect(comp.maxRendered < dir.contents.length).toBe(true, "If we have the same max as total contents this does nothing");
        expect($('.preview-img', el).length).toBe(comp.maxRendered, "Should render the last selected item, plus 1 previous");
    });


});

