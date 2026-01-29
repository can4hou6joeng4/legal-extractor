using System;
using System.IO;
using System.Runtime.InteropServices.WindowsRuntime;
using System.Threading.Tasks;
using Windows.Data.Pdf;
using Windows.Graphics.Imaging;
using Windows.Media.Ocr;
using Windows.Storage.Streams;

class Program
{
    static async Task<int> Main(string[] args)
    {
        if (args.Length < 2)
        {
            Console.WriteLine("Usage: WinOcrBridge <pdfPath> <pageNumber>");
            return 1;
        }

        string pdfPath = args[0];
        if (!int.TryParse(args[1], out int pageNumber) || pageNumber < 1)
        {
            Console.WriteLine("Invalid page number.");
            return 1;
        }

        try
        {
            string text = await RecognizePdfPageAsync(pdfPath, pageNumber);
            Console.WriteLine(text);
            return 0;
        }
        catch (Exception ex)
        {
            Console.Error.WriteLine($"Error: {ex.Message}");
            return 1;
        }
    }

    static async Task<string> RecognizePdfPageAsync(string pdfPath, int pageNumber)
    {
        // 1. 加载 PDF 文档
        var file = await Windows.Storage.StorageFile.GetFileFromPathAsync(Path.GetFullPath(pdfPath));
        var pdfDoc = await PdfDocument.LoadFromFileAsync(file);

        if (pageNumber > pdfDoc.PageCount)
            throw new Exception("Page number out of range.");

        // 2. 获取并渲染指定页面
        using var page = pdfDoc.GetPage((uint)pageNumber - 1);
        using var stream = new InMemoryRandomAccessStream();

        // 渲染为高分辨率位图以提高 OCR 准确率 (DPI 设为 300)
        var options = new PdfPageRenderOptions { DestinationWidth = (uint)(page.Size.Width * 3) };
        await page.RenderToStreamAsync(stream, options);

        // 3. 解码图像数据
        var decoder = await BitmapDecoder.CreateAsync(stream);
        using var softwareBitmap = await decoder.GetSoftwareBitmapAsync();

        // 4. 执行系统原生 OCR
        // 优先使用中文简体，如果没装则使用系统默认
        var lang = new Windows.Globalization.Language("zh-Hans-CN");
        var engine = OcrEngine.IsLanguageSupported(lang)
            ? OcrEngine.TryCreateFromLanguage(lang)
            : OcrEngine.TryCreateFromUserProfileLanguages();

        if (engine == null)
            throw new Exception("OCR Engine could not be initialized.");

        var result = await engine.RecognizeAsync(softwareBitmap);
        return result.Text;
    }
}
