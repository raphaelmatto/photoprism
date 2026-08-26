import { beforeEach, describe, expect, it, vi } from "vitest";
import PPageAlbumPhotos from "page/album/photos.vue";

describe("page/album/photos", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("applies the album date direction before the initial search", async () => {
    const album = { Order: "oldest", Title: "Folder" };
    const context = {
      uid: "atfolder00000000",
      model: { find: vi.fn().mockResolvedValue(album) },
      filter: { reverse: false },
      $route: { query: {} },
      $config: { get: vi.fn().mockReturnValue("PhotoPrism") },
    };

    context.sortReverse = () => PPageAlbumPhotos.methods.sortReverse.call(context);

    await PPageAlbumPhotos.methods.findAlbum.call(context);

    expect(context.model).toBe(album);
    expect(context.filter.reverse).toBe(true);
  });
});
